// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root

package goap

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

const linearCutoff = 0 // 1 cache line

var pool = sync.Pool{
	New: func() any {
		return &State{
			hx: 0,
			vx: make([]rule, 0, 16),
		}
	},
}

func newState(capacity int) *State {
	state := pool.Get().(*State)
	if cap(state.vx) < capacity {
		state.vx = make([]rule, 0, capacity)
	}
	return state
}

// ------------------------------------ State ------------------------------------

// State represents a state of the world.
type State struct {
	hx uint32 // Hash of the state
	vx []rule // Keys and values, interleaved
	node
}

type node struct {
	action    Action  // The action that led to this state
	parent    *State  // Pointer to the parent state
	heuristic float32 // Heuristic cost from this state to the goal
	stateCost float32 // Cost from the start state to this state
	totalCost float32 // Sum of cost and heuristic
	index     int     // Index of the state in the heap
	visited   bool    // Whether the state was visited
}

// StateOf creates a new state from a list of keys.
func StateOf(rules ...string) *State {
	state := newState(len(rules))
	for _, fact := range rules {
		if err := state.Add(fact); err != nil {
			panic(err)
		}
	}
	return state
}

func (s *State) release() {
	clear(s.vx)
	s.hx = 0
	s.vx = s.vx[:0]
	s.node = node{}
	pool.Put(s)
}

func (s *State) sort() {
	sort.Sort(s)
}

func (s *State) findLinear(key fact) (int, bool) {
	for i := 0; i < len(s.vx); i++ {
		if key == s.vx[i].Fact() {
			return i, true
		}
	}
	return 0, false
}

func (s *State) find(key fact) (int, bool) {
	if len(s.vx) <= linearCutoff {
		return s.findLinear(key)
	}

	x := sort.Search(len(s.vx), func(i int) bool { return s.vx[i].Fact() <= key })
	if x < len(s.vx) && s.vx[x].Fact() == key {
		return x, true
	}
	return x, false
}

// Store stores a key in the state, note that it requires rehashing the state
// and sorting the keys. This is NOT DONE by this method. The return value
// indicates whether the key was added to the state (true) or updated (false).
func (s *State) store(k fact, v expr) {
	r := ruleOf(k, v)

	// Check if the key already exists
	if i, ok := s.find(k); ok {
		s.hx ^= s.vx[i].Hash()
		s.hx ^= r.Hash()
		s.vx[i] = r
		return
	}

	// If not, add it to the state
	s.hx ^= r.Hash()
	s.vx = append(s.vx, ruleOf(k, v))
	s.sort()
}

// Add adds a key to the state.
func (s *State) Add(rule string) error {
	k, v, err := parseRule(rule)
	if err != nil {
		return err
	}

	s.store(k, v)
	return nil
}

// Del removes a key from the state.
func (s *State) Del(rule string) error {
	k, _, err := parseRule(rule)
	if err != nil {
		return err
	}

	i, ok := s.find(k)
	if !ok {
		return nil
	}

	// If we deleted, we need to sort and rehash. The sorting will place
	// the zero value at the end of the slice, so we can just trim it.
	s.hx ^= s.vx[i].Hash()
	s.vx[i] = 0
	s.sort()
	s.vx = s.vx[:len(s.vx)-1]
	return nil
}

func (s State) load(f fact) expr {
	if i, ok := s.find(f); ok {
		return s.vx[i].Expr()
	}
	return exprOf(opEqual, 0)
}

// Match checks if the State satisfies all the rules of the other state.
func (state *State) Match(needs *State) (bool, error) {
	i, j := 0, 0
	for i < len(needs.vx) && j < len(state.vx) {
		f0 := needs.vx[i].Fact()
		f1 := state.vx[j].Fact()

		switch {
		case f1 == f0:
			e0 := needs.vx[i].Expr()
			e1 := state.vx[j].Expr()

			if e1.Operator() != opEqual {
				return false, fmt.Errorf("plan: cannot match '%s%s', invalid state '%s'",
					f1.String(), e0.String(), e1.String())
			}

			match := false
			switch e0.Operator() {
			case opEqual:
				match = e1.Value() == e0.Value()
			case opLess:
				match = e1.Value() < e0.Value()
			case opGreater:
				match = e1.Value() > e0.Value()
			default:
				return false, fmt.Errorf("plan: cannot match '%s%s', invalid operator '%s'",
					f1.String(), e0.String(), e0.Operator().String())
			}

			if !match {
				return false, nil
			}

			j++
			i++
		case f1 > f0:
			j++
		default: // No match
			return false, nil
		}
	}

	// Check if all elements of other.vx were matched
	return i == len(needs.vx), nil
}

// Apply adds (applies) the keys from the effects to the state.
func (s *State) Apply(effects *State) error {
	for _, elem := range effects.vx {
		f, e := elem.Fact(), elem.Expr()
		x := s.load(f)

		// Current state must be a full state
		if x.Operator() != opEqual {
			return fmt.Errorf("plan: cannot apply '%s%s', invalid state '%s'", f.String(), e.String(), x.String())
		}

		// Apply the effect to the state
		switch e.Operator() {
		case opEqual:
			s.store(f, e)
		case opIncrement:
			s.store(f, exprOf(x.Operator(), x.Percent()+e.Percent()))
		case opDecrement:
			s.store(f, exprOf(x.Operator(), x.Percent()-e.Percent()))
		default:
			return fmt.Errorf("plan: cannot apply '%s%s', invalid predict operator '%s'", f.String(), e.String(), e.Operator().String())
		}
	}

	return nil
}

// Distance estimates the distance to the goal state as the number of differing keys.
func (s *State) Distance(goal *State) (diff float32) {
	for _, elem := range goal.vx {
		k, v := elem.Fact(), elem.Expr()
		y := expr(v).Percent()
		x := s.load(fact(k)).Percent()
		switch {
		case x > y:
			diff += x - y
		case x < y:
			diff += y - x
		}
	}
	return diff
}

// Equals returns true if the state is equal to the other state.
func (s *State) Equals(other *State) bool {
	return s.Hash() == other.Hash()
}

// Hash returns a hash of the state.
func (s *State) Hash() (h uint32) {
	return s.hx
}

// Clone returns a clone of the state.
func (s *State) Clone() *State {
	clone := newState(len(s.vx))
	clone.hx = s.hx
	clone.vx = clone.vx[:len(s.vx)]
	copy(clone.vx, s.vx)
	return clone
}

// String returns a string representation of the state.
func (s *State) String() string {
	values := make([]string, 0, len(s.vx))
	for _, elem := range s.vx {
		values = append(values, elem.Fact().String()+elem.Expr().String())
	}

	return "{" + strings.Join(values, ", ") + "}"
}

// Len returns the number of elements in the state.
func (s *State) Len() int {
	return len(s.vx)
}

// Less reports whether the element with index i should sort before the element with index j.
func (s *State) Less(i, j int) bool {
	return s.vx[i].Fact() > s.vx[j].Fact()
}

// Swap swaps the elements with indexes i and j.
func (s *State) Swap(i, j int) {
	s.vx[i], s.vx[j] = s.vx[j], s.vx[i]
}
