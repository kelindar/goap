// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root

package goap

import (
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/zeebo/xxh3"
)

var (
	factRegex = regexp.MustCompile(`^(!*)([a-zA-Z_]+)$`)
	factCache = new(sync.Map)
)

// fact represents a state fact.
type fact uint64

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
func factOf(s string) (fact, bool) {
	matches := factRegex.FindStringSubmatch(s)
	if len(matches) != 3 {
		panic("invalid fact: " + s)
	}

	// Hash the fact and cache it for reverse lookup
	f := fact(xxh3.HashString(strings.ToLower(matches[2])))
	v := matches[1] == ""

	factCache.Store(f, matches[2])
	return f, v
}

// State represents a state of the world.
type State map[fact]bool

// StateOf creates a new state from a list of keys.
func StateOf(facts ...string) State {
	state := make(State, len(facts))
	for _, fact := range facts {
		state.Add(fact)
	}
	return state
}

// Add adds a key to the state.
func (s *State) Add(fact string) {
	k, v := factOf(fact)
	(*s)[k] = v
}

// Remove removes a key from the state.
func (s *State) Remove(fact string) {
	k, _ := factOf(fact)
	delete(*s, k)
}

// has returns true if the state contains the fact with a given state.
func (s *State) has(f fact, v bool) bool {
	if val, ok := (*s)[f]; ok {
		return val == v
	}
	return false
}

// Has checks if the State contains all the keys from another State.
func (s State) Has(other State) bool {
	for f, v := range other {
		if !s.has(f, v) {
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

	return s.Hash() == other.Hash()
}

// Clone returns a clone of the state.
func (s *State) Clone() State {
	clone := make(State, len(*s))
	for k, v := range *s {
		clone[k] = v
	}
	return clone
}

// Apply adds (applies) the keys from the effects to the state.
func (s *State) Apply(effects State) {
	for k, v := range effects {
		(*s)[k] = v
	}
}

// Distance estimates the distance to the goal state as the number of differing keys.
func (s *State) Distance(goal State) (diff float32) {
	for f, v := range goal {
		if !s.has(f, v) {
			diff++
		}
	}
	return diff
}

// Hash returns a hash of the state.
func (s *State) Hash() (h uint64) {
	for k := range *s {
		h ^= uint64(k)
	}
	return
}

// String returns a string representation of the state.
func (s *State) String() string {
	values := make([]string, 0, len(*s))
	for k, v := range *s {
		if v {
			values = append(values, k.String())
		} else {
			values = append(values, "!"+k.String())
		}
	}

	sort.Strings(values)
	return "{" + strings.Join(values, ", ") + "}"
}
