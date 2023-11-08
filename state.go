// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root

package goap

import (
	"fmt"
	"io"
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

// parseRule parses an expression containing a fact and a rule
func parseRule(s string) (fact, expr, error) {
	length := len(s)
	if length == 0 {
		return 0, 0, fmt.Errorf("plan: rule is an empty string")
	}

	key := [2]int{0, 0}   // [start, end]
	value := float64(100) // default value
	op := opEqual         // default operator

	var i int
	var valueStr string

	// Check for initial '!'
	if s[0] == '!' {
		if length == 1 {
			return 0, 0, fmt.Errorf("plan: invalid rule '%s'", s)
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

	return factOf(s[key[0]:i]), exprOf(opEqual, value), nil

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
		return 0, 0, fmt.Errorf("plan: invalid operator '%c' in rule '%s'", s[i], s)
	}

	i++
	valueStr = s[i:]

	// Parse the floating-point value
	value, err := strconv.ParseFloat(valueStr, 32)
	if err != nil || value < valueMin || value > valueMax {
		return 0, 0, fmt.Errorf("plan: invalid value '%s' in rule '%s'", valueStr, s)
	}

	return factOf(s[key[0]:key[1]]), exprOf(op, value), nil
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

// expr represents an expression, expressed as a fixed point between 0 and 100.00,
// the value can also be a delta (+/-) from the current value or a comparison operator
// first 4 bits are used to indicate the type of the expr (operator).
// [0-3]  - operator
// [4-15] - unused
// [16-31] - value
type expr uint32

// exprOf creates a new expression from an operator and a value.
func exprOf(op operator, value float64) expr {
	return expr(uint32(op)<<28 | uint32(value*100))
}

// Operator returns the operator of the effect.
func (e expr) Operator() operator {
	return operator(e >> 28)
}

// Value returns the value of the effect.
func (e expr) Value() uint32 {
	return uint32(e & 0xFFFF)
}

// Percent returns the value as a percentage.
func (e expr) Percent() float32 {
	if e.Value() >= valueMax {
		return 100
	}
	return float32(e.Value()) / 100
}

// String returns the string representation of the effect.
func (e expr) String() string {
	return e.Operator().String() + strconv.FormatFloat(float64(e.Percent()), 'f', 2, 32)
}

// ------------------------------------ State ------------------------------------

// State represents a state of the world.
type State struct {
	m *intmap.Map
}

// StateOf creates a new state from a list of keys.
func StateOf(rules ...string) State {
	state := State{m: intmap.New(8, 0.9)}
	for _, fact := range rules {
		state.Add(fact)
	}
	return state
}

// Add adds a key to the state.
func (s State) Add(rule string) error {
	k, v, err := parseRule(rule)
	if err != nil {
		return err
	}

	s.m.Store(uint32(k), uint32(v))
	return nil
}

// Remove removes a key from the state.
func (s State) Remove(rule string) error {
	k, _, err := parseRule(rule)
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

// Match checks if the State satisfies all the rules of the other state.
func (s State) Match(other State) (bool, error) {
	match := true
	err := other.m.RangeErr(func(k, v uint32) error {
		f, e := fact(k), expr(v)
		x := s.load(f)

		// Current state must be a full state
		if x.Operator() != opEqual {
			return fmt.Errorf("plan: cannot satisfy '%s%s', invalid state '%s'", f.String(), e.String(), x.String())
		}

		// Check if the state satisfies the rule
		switch e.Operator() {
		case opEqual:
			match = x.Value() == e.Value()
		case opLess:
			match = x.Value() <= e.Value()
		case opGreater:
			match = x.Value() >= e.Value()
		default:
			return fmt.Errorf("plan: cannot satisfy '%s%s', invalid operator '%s'", f.String(), e.String(), e.Operator().String())
		}

		// Short-circuit if the state doesn't match
		if !match {
			return io.EOF
		}

		return nil
	})

	switch err {
	case io.EOF:
		return match, nil
	default:
		return match, err
	}
}

func (s State) load(f fact) expr {
	v, ok := s.m.Load(uint32(f))
	if !ok {
		return exprOf(opEqual, 0)
	}
	return expr(v)
}

// Apply adds (applies) the keys from the effects to the state.
func (s State) Apply(effects State) error {
	return effects.m.RangeErr(func(k, v uint32) error {
		f, e := fact(k), expr(v)
		x := s.load(f)

		// Current state must be a full state
		if x.Operator() != opEqual {
			return fmt.Errorf("plan: cannot apply '%s%s', invalid state '%s'", f.String(), e.String(), x.String())
		}

		// Apply the effect to the state
		switch e.Operator() {
		case opEqual:
			s.m.Store(k, e.Value())
		case opIncrement:
			s.m.Store(k, x.Value()+e.Value())
		case opDecrement:
			s.m.Store(k, x.Value()-e.Value())
		default:
			return fmt.Errorf("plan: cannot apply '%s%s', invalid predict operator '%s'", f.String(), e.String(), e.Operator().String())
		}
		return nil
	})
}

// Distance estimates the distance to the goal state as the number of differing keys.
func (s State) Distance(goal State) (diff float32) {
	goal.m.RangeEach(func(k, v uint32) {
		if !s.has(fact(k), v) {
			diff++
		}
	})
	return diff
}

// Equals returns true if the state is equal to the other state.
func (s State) Equals(other State) bool {
	if s.m.Count() != other.m.Count() {
		return false
	}
	return s.Hash() == other.Hash()
}

// Hash returns a hash of the state.
func (s State) Hash() (h uint64) {
	s.m.Range(func(k, v uint32) bool {
		h ^= (uint64(k) << 32) | uint64(v)
		return true
	})
	return
}

// Clone returns a clone of the state.
func (s State) Clone() State {
	return State{
		m: s.m.Clone(),
	}
}

// String returns a string representation of the state.
func (s State) String() string {
	values := make([]string, 0, s.m.Count())
	s.m.RangeEach(func(k, v uint32) {
		values = append(values, fact(k).String()+expr(v).String())
	})

	sort.Strings(values)
	return "{" + strings.Join(values, ", ") + "}"
}
