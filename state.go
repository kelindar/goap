// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root

package goap

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/kelindar/intmap"
	"github.com/zeebo/xxh3"
)

var (
	factRegex = regexp.MustCompile(`^(!*)([a-zA-Z_]+)$`)
	factCache = new(sync.Map)
)

// ------------------------------------ Fact ------------------------------------

// fact represents a state fact.
type fact uint32

// factOf creates a new fact from a string.
func factOf(s string) fact {
	f := fact(xxh3.HashString(strings.ToLower(s)))
	factCache.Store(f, s)
	return f
}

// String returns the string representation of the fact.
func (f fact) String() string {
	if v, ok := factCache.Load(f); ok {
		return v.(string)
	}
	return "unknown"
}

// parseExpr parses an expression containing a fact and a rule
func parseExpr(s string) (fact, rule, error) {
	length := len(s)
	if length == 0 {
		return 0, 0, fmt.Errorf("parse: empty string")
	}

	key := [2]int{0, 0}   // [start, end]
	value := float64(100) // default value
	op := opEqual         // default operator

	var i int
	var valueStr string

	// Check for initial '!'
	if s[0] == '!' {
		if length == 1 {
			return 0, 0, fmt.Errorf("parse: '!' found without following characters")
		}

		op = opEqual
		value = float64(0)
		valueStr = "0"
		key[0] = 1
		i = 1
		goto parseKey
	}

	// Parse the key in the form of [a-zA-Z_]+
parseKey:
	for ; i < length; i++ {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' {
			continue
		}
		key[1] = i
		goto parseOperator
	}

	return factOf(s[key[0]:i]), ruleOf(opEqual, value), nil

	// Parse the operator in the form of [=+-<>]
parseOperator:
	switch s[i] {
	case '=':
		op = opEqual
	case '+':
		op = opIncrement
	case '-':
		op = opDecrement
	case '<':
		op = opLess
	case '>':
		op = opGreater
	default:
		return 0, 0, fmt.Errorf("parse: invalid operator '%c'", s[i])
	}

	i++
	valueStr = s[i:]

	// Parse the floating-point value
	value, err := strconv.ParseFloat(valueStr, 32)
	if err != nil || value < valueMin || value > valueMax {
		return 0, 0, fmt.Errorf("parse: invalid value '%s'", valueStr)
	}

	return factOf(s[key[0]:key[1]]), ruleOf(op, value), nil
}

// ------------------------------------ Effect ------------------------------------

const (
	valueMin = 0
	valueMax = 10000 // 100.00 is the max value for a percentage
)

const (
	opEqual operator = iota
	opIncrement
	opDecrement
	opLess
	opGreater
)

type operator uint32

// String returns the string representation of the operator.
func (o operator) String() string {
	switch o {
	case opIncrement:
		return "+"
	case opDecrement:
		return "-"
	case opLess:
		return "<"
	case opGreater:
		return ">"
	case opEqual:
		fallthrough
	default:
		return "="
	}
}

// rule represents a fact rule, expressed as a fixed point between 0 and 100.00,
// the value can also be a delta (+/-) from the current value or a comparison operator
// first 4 bits are used to indicate the type of the rule (operator).
// [0-3]  - operator
// [4-15] - unused
// [16-31] - value
type rule uint32

// ruleOf creates a new effect from an operator and a value.
func ruleOf(op operator, value float64) rule {
	return rule(uint32(op)<<28 | uint32(value*100))
}

// Operator returns the operator of the effect.
func (e rule) Operator() operator {
	return operator(e >> 28)
}

// Value returns the value of the effect.
func (e rule) Value() uint32 {
	return uint32(e & 0xFFFF)
}

// Percent returns the value as a percentage.
func (e rule) Percent() float32 {
	if e.Value() >= valueMax {
		return 100
	}
	return float32(e.Value()) / 100
}

// String returns the string representation of the effect.
func (e rule) String() string {
	return e.Operator().String() + strconv.FormatFloat(float64(e.Percent()), 'f', 2, 32)
}

// ------------------------------------ State ------------------------------------

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
func (s State) Add(fact string) error {
	k, v, err := parseExpr(fact)
	if err != nil {
		return err
	}

	s.m.Store(uint32(k), uint32(v))
	return nil
}

// Remove removes a key from the state.
func (s State) Remove(fact string) error {
	k, _, err := parseExpr(fact)
	if err != nil {
		return err
	}

	s.m.Delete(uint32(k))
	return nil
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
