package goap

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/zeebo/xxh3"
)

var factCache = new(sync.Map)

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
	value := float32(100) // default value
	op := opEqual         // default operator

	var i int
	var valueStr string

	// Check for initial '!'
	if s[0] == '!' {
		if length == 1 {
			return 0, 0, fmt.Errorf("plan: invalid rule '%s'", s)
		}

		op = opEqual
		value = float32(0)
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
	val, err := strconv.ParseFloat(valueStr, 32)
	if err != nil || value < valueMin || value > valueMax {
		return 0, 0, fmt.Errorf("plan: invalid value '%s' in rule '%s'", valueStr, s)
	}

	return factOf(s[key[0]:key[1]]), exprOf(op, float32(val)), nil
}

// ------------------------------------ Expression ------------------------------------

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
func exprOf(op operator, value float32) expr {
	if value < 0 {
		value = 0
	}
	if value > 100 {
		value = 100
	}
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

// ------------------------------------ Packed Data ------------------------------------

type rule uint64

func elemOf(f fact, e expr) rule {
	return rule(f)<<32 | rule(e)
}

func (e rule) Fact() fact {
	return fact(e >> 32)
}

func (e rule) Expr() expr {
	return expr(e & 0xFFFFFFFF)
}
