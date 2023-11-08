// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root

package goap

import (
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/kelindar/intmap"
	"github.com/zeebo/xxh3"
)

var (
	factRegex = regexp.MustCompile(`^(!*)([a-zA-Z_]+)$`)
	factCache = new(sync.Map)
)

const (
	minValue = 0
	maxValue = 100
)

// fact represents a state fact.
type fact uint32

// String returns the string representation of the fact.
func (f fact) String() string {
	if v, ok := factCache.Load(f); ok {
		return v.(string)
	}
	return "unknown"
}

// factOf parses a state fact from a string. It returns the hash of the fact and
// a boolean indicating the presence or absence of the fact. The format of the fact
// is the following: "key" or "!key" where the exclamation mark indicates the
// absence of the fact.
func factOf(s string) (fact, uint32) {
	matches := factRegex.FindStringSubmatch(s)
	if len(matches) != 3 {
		panic("invalid fact: " + s)
	}

	// Hash the fact and cache it for reverse lookup
	f := fact(xxh3.HashString(strings.ToLower(matches[2])))
	v := uint32(maxValue)
	if matches[1] == "!" {
		v = minValue
	}

	factCache.Store(f, matches[2])
	return f, v
}

// State represents a state of the world.
type State struct {
	m *intmap.Map
}

// StateOf creates a new state from a list of keys.
func StateOf(facts ...string) State {
	state := State{m: intmap.New(8, 0.9)}
	for _, fact := range facts {
		state.Add(fact)
	}
	return state
}

// Add adds a key to the state.
func (s State) Add(fact string) {
	k, v := factOf(fact)
	s.m.Store(uint32(k), v)
}

// Remove removes a key from the state.
func (s State) Remove(fact string) {
	k, _ := factOf(fact)
	s.m.Delete(uint32(k))
}

// has returns true if the state contains the fact with a given state.
func (s State) has(f fact, x uint32) bool {
	v, ok := s.m.Load(uint32(f))
	return ok && v >= x
}

// Has checks if the State contains all the keys from another State.
func (s State) Has(other State) (ok bool) {
	ok = true
	other.m.Range(func(k, v uint32) bool {
		if !s.has(fact(k), v) {
			ok = false
			return false
		}
		return true
	})
	return
}

// Equals returns true if the state is equal to the other state.
func (s State) Equals(other State) bool {
	if s.m.Count() != other.m.Count() {
		return false
	}
	return s.Hash() == other.Hash()
}

// Clone returns a clone of the state.
func (s State) Clone() State {
	return State{
		m: s.m.Clone(),
	}
}

// Apply adds (applies) the keys from the effects to the state.
func (s State) Apply(effects State) {
	effects.m.Range(func(k, v uint32) bool {
		s.m.Store(k, v)
		return true
	})
}

// Distance estimates the distance to the goal state as the number of differing keys.
func (s State) Distance(goal State) (diff float32) {
	goal.m.Range(func(k, v uint32) bool {
		if !s.has(fact(k), v) {
			diff++
		}
		return true
	})
	return diff
}

// Hash returns a hash of the state.
func (s State) Hash() (h uint64) {
	s.m.Range(func(k, v uint32) bool {
		h ^= (uint64(k) << 32) | uint64(v)
		return true
	})
	return
}

// String returns a string representation of the state.
func (s State) String() string {
	values := make([]string, 0, s.m.Count())
	s.m.Range(func(k, v uint32) bool {
		switch {
		case v == 0:
			values = append(values, "!"+fact(k).String())
		default:
			values = append(values, fact(k).String())
		}
		return true
	})

	sort.Strings(values)
	return "{" + strings.Join(values, ", ") + "}"
}
