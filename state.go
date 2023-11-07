// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root

package goap

import (
	"sort"
	"strings"

	"github.com/zeebo/xxh3"
)

// State represents a state of the world.
type State map[string]struct{}

// StateOf creates a new state from a list of keys.
func StateOf(s ...string) State {
	state := make(State)
	for _, v := range s {
		state[v] = struct{}{}
	}
	return state
}

// Add adds a key to the state.
func (s *State) Add(key string) {
	(*s)[key] = struct{}{}
}

// Remove removes a key from the state.
func (s *State) Remove(key string) {
	delete(*s, key)
}

// Has returns true if the state contains the key.
func (s *State) Has(key string) bool {
	_, ok := (*s)[key]
	return ok
}

// HasAll checks if the State contains all the keys from another State.
func (s State) HasAll(other State) bool {
	for k := range other {
		if !s.Has(k) {
			return false
		}
	}
	return true
}

// Equals returns true if the state is equal to the other state.
func (s *State) Equals(other State) bool {
	if len(*s) != len(other) {
		return false
	}

	for k := range *s {
		if _, ok := other[k]; !ok {
			return false
		}
	}
	return true
}

// Clone returns a clone of the state.
func (s *State) Clone() State {
	clone := make(State)
	for k := range *s {
		clone[k] = struct{}{}
	}
	return clone
}

// Apply adds (applies) the keys from the effects to the state.
func (s *State) Apply(effects State) {
	for k := range effects {
		s.Add(k)
	}

	// Remove keys that are not in effects
	/*for k := range *s {
		if _, ok := effects[k]; !ok {
			s.Remove(k)
		}
	}*/
}

// Distance estimates the distance to the goal state as the number of differing keys.
func (s *State) Distance(goal State) (diff float32) {
	for k := range goal {
		if !s.Has(k) {
			diff++
		}
	}
	return diff
}

// Hash returns a hash of the state.
func (s *State) Hash() (h uint64) {
	for k := range *s {
		h ^= xxh3.HashString(k)
	}
	return
}

// String returns a string representation of the state.
func (s *State) String() string {
	values := make([]string, 0, len(*s))
	for k := range *s {
		values = append(values, k)
	}
	sort.Strings(values)
	return "{" + strings.Join(values, ", ") + "}"
}
