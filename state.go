// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root

package goap

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
)

var (
	factRegex    = regexp.MustCompile(`^(!*)([a-zA-Z_]+)$`)
	factCache    = new(sync.Map)
	linearCutoff = 8 // 1 cache line
)

var states = sync.Pool{
	New: func() interface{} {
		return &State{
			hx: 0,
			vx: make([]rule, 0, 16),
		}
	},
}

func newState(capacity int) *State {
	state := states.Get().(*State)
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
}

// StateOf creates a new state from a list of keys.
func StateOf(rules ...string) *State {
	state := newState(len(rules))
	for _, fact := range rules {
		state.Add(fact)
	}
	state.rehash()
	return state
}

// rehash rehashes the state
func (s *State) rehash() {
	s.hx = 0
	for _, v := range s.vx {
		//s.h ^= (uint64(k) << 32) | (uint64(v)*0xdeece66d + 0xb)
		s.hx ^= (uint32(v.Fact()) | (uint32(v.Expr())*0xdeece66d + 0xb))
	}
}

func (s *State) release() {
	for i := range s.vx {
		s.vx[i] = 0
	}

	s.hx = 0
	s.vx = s.vx[:0]
	states.Put(s)
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

	x := sort.Search(len(s.vx), func(i int) bool { return s.vx[i].Fact() >= key })
	if x < len(s.vx) && s.vx[x].Fact() == key {
		return x, true
	}
	return x, false
}

func (s *State) sort() {
	if len(s.vx) > linearCutoff {
		sort.Slice(s.vx, func(i, j int) bool { return s.vx[i].Fact() < s.vx[j].Fact() })
	}
}

// Store stores a key in the state, note that it requires rehashing the state
// and sorting the keys. This is NOT DONE by this method. The return value
// indicates whether the key was added to the state (true) or updated (false).
func (s *State) store(k fact, v expr) bool {
	if i, ok := s.find(k); ok {
		s.vx[i] = elemOf(k, v)
		return false
	}

	// If not, add it to the state
	s.vx = append(s.vx, elemOf(k, v))
	return true
}

// Add adds a key to the state.
func (s *State) Add(rule string) error {
	k, v, err := parseRule(rule)
	if err != nil {
		return err
	}

	if added := s.store(k, v); added {
		s.sort()
	}

	s.rehash()
	return nil
}

// Remove removes a key from the state.
func (s *State) Remove(rule string) error {
	k, _, err := parseRule(rule)
	if err != nil {
		return err
	}

	i, ok := s.find(k)
	if !ok {
		return nil
	}

	// TODO: sort inverse, trim tail
	s.vx[i] = 0
	s.sort()
	s.rehash()
	return nil
}

func (s State) load(f fact) expr {
	if i, ok := s.find(f); ok {
		return s.vx[i].Expr()
	}
	return exprOf(opEqual, 0)
}

// Match checks if the State satisfies all the rules of the other state.
func (s *State) Match(other *State) (bool, error) {
	match := true
	for _, elem := range other.vx {
		f, e := elem.Fact(), elem.Expr()
		x := s.load(f)

		// Current state must be a full state
		if x.Operator() != opEqual {
			return false, fmt.Errorf("plan: cannot match '%s%s', invalid state '%s'", f.String(), e.String(), x.String())
		}

		// Check if the state satisfies the rule
		switch e.Operator() {
		case opEqual:
			match = x.Value() == e.Value()
		case opLess:
			match = x.Value() < e.Value()
		case opGreater:
			match = x.Value() > e.Value()
		default:
			return false, fmt.Errorf("plan: cannot match '%s%s', invalid operator '%s'", f.String(), e.String(), e.Operator().String())
		}

		// Short-circuit if the state doesn't match
		if !match {
			return false, nil
		}
	}
	return match, nil
}

// Apply adds (applies) the keys from the effects to the state.
func (s *State) Apply(effects *State) error {
	defer s.rehash()
	defer s.sort()

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

	sort.Strings(values)
	return "{" + strings.Join(values, ", ") + "}"
}
