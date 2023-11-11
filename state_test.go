// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root

package goap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

/*
cpu: 13th Gen Intel(R) Core(TM) i7-13700K
BenchmarkState/clone-24         	13685818	        87.07 ns/op	     224 B/op	       2 allocs/op
BenchmarkState/match-24         	314971872	        3.827 ns/op	       0 B/op	       0 allocs/op
BenchmarkState/add-24           	13094532	        87.17 ns/op	      40 B/op	       4 allocs/op
BenchmarkState/remove-24        	14003313	        85.90 ns/op	      40 B/op	       4 allocs/op
BenchmarkState/apply-24         	82659428	        15.14 ns/op	       0 B/op	       0 allocs/op
BenchmarkState/distance-24      	180420306	        6.620 ns/op	       0 B/op	       0 allocs/op
*/
func BenchmarkState(b *testing.B) {
	b.ReportAllocs()

	b.Run("clone", func(b *testing.B) {
		s1, s2 := StateOf("A", "B", "C", "D"), StateOf()
		for i := 0; i < b.N; i++ {
			s2 = s1.Clone()
		}
		assert.NotNil(b, s2)
	})

	b.Run("match", func(b *testing.B) {
		state1 := StateOf("A", "B", "C")
		state2 := StateOf("A", "B")
		for i := 0; i < b.N; i++ {
			state1.Match(state2)
		}
	})

	b.Run("add", func(b *testing.B) {
		state := StateOf()
		for i := 0; i < b.N; i++ {
			state.Add("A")
		}
	})

	b.Run("remove", func(b *testing.B) {
		state := StateOf()
		state.Add("A")
		for i := 0; i < b.N; i++ {
			state.Del("A")
		}
	})

	b.Run("apply", func(b *testing.B) {
		state1 := StateOf("A", "B", "C")
		state2 := StateOf("D", "E")
		for i := 0; i < b.N; i++ {
			state1.Apply(state2)
		}
	})

	b.Run("distance", func(b *testing.B) {
		state1 := StateOf("A", "B", "C")
		state2 := StateOf("A", "B", "D")
		for i := 0; i < b.N; i++ {
			state1.Distance(state2)
		}
	})
}

func TestMatchSimple(t *testing.T) {
	state1 := StateOf("A", "B", "C")
	state2 := StateOf("A", "B")

	// Must be sorted
	assert.Equal(t, "{C=100, B=100, A=100}", state1.String())
	assert.Equal(t, "{B=100, A=100}", state2.String())

	ok1, err := state1.Match(state2)
	assert.NoError(t, err)
	assert.True(t, ok1)

	ok2, err := state2.Match(state1)
	assert.NoError(t, err)
	assert.False(t, ok2)
}

func TestMatchNumeric(t *testing.T) {
	state1 := StateOf("A=50", "B=100")
	state2 := StateOf("A>10", "B=100")

	ok, err := state1.Match(state2)
	assert.NoError(t, err)
	assert.True(t, ok)

	_, err = state2.Match(state1)
	assert.Error(t, err)
}

func TestHash(t *testing.T) {
	state1 := StateOf("A", "B", "C")
	state2 := StateOf("C", "B", "A")
	state3 := StateOf("A", "B", "C", "D")

	assert.Equal(t, state1.Hash(), state2.Hash())
	assert.NotEqual(t, state1.Hash(), state3.Hash())
	assert.NotEqual(t, state2.Hash(), state3.Hash())
}

func TestNumericHash(t *testing.T) {
	state1 := StateOf("food=0", "hunger=0", "tired=0")
	state2 := StateOf("food=10", "hunger=0", "tired=10")

	assert.NotEqual(t, state1.Hash(), state2.Hash())
}

func TestStateEquals(t *testing.T) {
	state1 := StateOf("A", "B", "C")
	state2 := StateOf("A", "B", "C")
	state3 := StateOf("A", "B", "C", "D")
	state4 := StateOf("A", "B")
	state5 := StateOf("A", "D")

	assert.True(t, state1.Equals(state2))
	assert.True(t, state2.Equals(state1))
	assert.False(t, state1.Equals(state3))
	assert.False(t, state3.Equals(state1))
	assert.False(t, state1.Equals(state4))
	assert.False(t, state4.Equals(state1))
	assert.False(t, state4.Equals(state5))
}

func TestClone(t *testing.T) {
	state := StateOf("A", "B", "C")
	clone := state.Clone()

	// Ensure the clone is equal to the original
	assert.True(t, clone.Equals(state))

	// Ensure the clone is a separate instance
	state.Del("A")
	assert.False(t, clone.Equals(state))
}

func TestDistance(t *testing.T) {
	tests := []struct {
		state1, state2 []string
		expect         float32
	}{
		{[]string{"A"}, []string{"A"}, 0},
		{[]string{"A=100"}, []string{"A=10"}, 90},
		{[]string{"A=100"}, []string{"A=90"}, 10},
		{[]string{"A"}, []string{"B"}, 200},
		{[]string{"A"}, []string{"A", "B"}, 100},
		{[]string{"A", "B"}, []string{"A"}, 100},
		{[]string{"A", "B"}, []string{"C", "D"}, 400},
		{[]string{"A", "B"}, []string{"A", "B"}, 0},
		{[]string{"A", "B"}, []string{"A", "B", "C"}, 100},
		{[]string{"A", "B", "C"}, []string{"D", "B"}, 300},
	}

	for _, test := range tests {
		state1 := StateOf(test.state1...)
		state2 := StateOf(test.state2...)
		assert.Equal(t, test.expect, state1.Distance(state2),
			"state1=%s, state2=%s", state1, state2)
	}
}

func TestStateString(t *testing.T) {
	state := StateOf("A", "B", "C")
	assert.Equal(t, "{C=100, B=100, A=100}", state.String())

	state = StateOf()
	assert.Equal(t, "{}", state.String())
}

func TestRemove(t *testing.T) {
	state := StateOf("A", "B", "C", "D", "E", "F", "G", "H", "I")

	state.Del("E")
	state.Del("F")
	assert.Equal(t, "{H=100, G=100, D=100, C=100, B=100, A=100, I=100}",
		state.String())
}

func TestApply(t *testing.T) {
	tests := []struct {
		state1, state2, expect []string
	}{
		{[]string{"A"}, []string{"A"}, []string{"A"}},
		{[]string{"A"}, []string{"A+10"}, []string{"A"}},
		{[]string{"A"}, []string{"A-10"}, []string{"A=90"}},
		{[]string{"A"}, []string{"B"}, []string{"A", "B"}},
		{[]string{"A"}, []string{"A", "B"}, []string{"A", "B"}},
		{[]string{"A", "B"}, []string{"A"}, []string{"A", "B"}},
		{[]string{"A", "B"}, []string{"B-10"}, []string{"A", "B=90"}},
		{[]string{"A", "B"}, []string{"C", "D"}, []string{"A", "B", "C", "D"}},
	}

	for _, test := range tests {
		state1 := StateOf(test.state1...)
		state2 := StateOf(test.state2...)
		state1.Apply(state2)

		expect := StateOf(test.expect...)
		ok, err := state1.Match(expect)
		assert.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, expect.String(), state1.String())
	}
}

func TestApplySort(t *testing.T) {
	state1 := StateOf("A", "B")
	state2 := StateOf("D")

	// Under 8 elements, should not sort
	state1.Apply(state2)
	assert.Equal(t, "{D=100, B=100, A=100}", state1.String())

	// Over 8 elements, should sort
	state3 := StateOf("D", "E", "F", "G", "H", "I", "J")
	state1.Apply(state3)
	assert.Equal(t, "{H=100, G=100, J=100, D=100, F=100, B=100, E=100, A=100, I=100}",
		state1.String())
}

func TestApplyError(t *testing.T) {
	state1 := StateOf("A>10")
	state2 := StateOf("A")
	assert.Error(t, state1.Apply(state2))
	assert.Error(t, state2.Apply(state1))
}
