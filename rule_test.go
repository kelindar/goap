// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root

package goap

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tests := map[string]string{
		"hp":         "hp=100.00",
		"!hp":        "hp=0.00",
		"hp=10":      "hp=10.00",
		"hp=10.5":    "hp=10.50",
		"hp=10.":     "hp=10.00",
		"hp-1":       "hp-1.00",
		"hp+1":       "hp+1.00",
		"hp+1.5":     "hp+1.50",
		"hp-1.5":     "hp-1.50",
		"hp=200":     "hp=100.00",
		"hp=0":       "hp=0.00",
		"hp=0.5":     "hp=0.50",
		"hp=0.":      "hp=0.00",
		"hp-0.0":     "hp-0.00",
		"hp>10":      "hp>10.00",
		"hp<10":      "hp<10.00",
		"ammo_max":   "ammo_max=100.00",
		"!ammo_max":  "ammo_max=0.00",
		"ammo_Max=0": "ammo_Max=0.00",
		"abc2":       "abc2=100.00",
		"hp>=10":     "(error)",
		"hp<=10":     "(error)",
		"hp 2":       "(error)",
		"hp=2.2.2":   "(error)",
		"hp ":        "(error)",
		"":           "(error)",
		"!":          "(error)",
	}

	for input, expect := range tests {
		k, v, err := parseRule(input)
		if expect == "(error)" {
			assert.Error(t, err)
			continue
		}

		assert.NoError(t, err)
		assert.Equal(t, expect, fmt.Sprintf("%s%s", k.String(), v.String()), input)
	}
}

func TestRuleHash(t *testing.T) {
	tests := []struct {
		rules  []string
		expect []string
	}{
		{[]string{"B"}, []string{"A", "C"}},
		{[]string{"B", "C"}, []string{"A"}},
		{[]string{"A", "A"}, []string{"A", "B", "C"}},
		{[]string{"A", "B", "C"}, []string{}},
		{[]string{"A", "B", "C", "D"}, []string{"D"}},
		{[]string{"A", "A=50"}, []string{"A=50", "B", "C"}},
		{[]string{"X1", "D"}, []string{"A", "B", "C", "X1", "D"}},
	}

	for _, test := range tests {
		state := hashOf("A", "B", "C")
		for _, rule := range test.rules {
			state ^= hashOf(rule)
		}

		assert.Equal(t, state, hashOf(test.expect...),
			strings.Join(test.rules, ", "))
	}
}

func TestFactString(t *testing.T) {
	assert.Equal(t, "A", factOf("A").String())
	assert.Equal(t, "unknown", fact(123).String())
}

// ------------------------------------ Test Functions ------------------------------------

func hashOf(s ...string) (h uint32) {
	for _, r := range s {
		k, v, err := parseRule(r)
		if err != nil {
			panic(err)
		}

		h ^= ruleOf(k, v).Hash()
	}
	return
}
