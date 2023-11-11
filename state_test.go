// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root

package goap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

/*
cpu: 13th Gen Intel(R) Core(TM) i7-13700K
BenchmarkState/clone-24         	13919708	        85.71 ns/op	     208 B/op	       2 allocs/op
BenchmarkState/match-24         	125247872	        9.597 ns/op	       0 B/op	       0 allocs/op
BenchmarkState/add-24           	14518809	        82.26 ns/op	      40 B/op	       4 allocs/op
BenchmarkState/remove-24        	14153847	        84.15 ns/op	      40 B/op	       4 allocs/op
BenchmarkState/apply-24         	65917405	        18.15 ns/op	       0 B/op	       0 allocs/op
BenchmarkState/distance-24      	72343632	        17.25 ns/op	       0 B/op	       0 allocs/op
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

func TestStateApply(t *testing.T) {
	state1 := StateOf("A", "B", "C")
	state2 := StateOf("D", "E")
	state1.Apply(state2)

	expect := StateOf("A", "B", "C", "D", "E")
	ok, err := state1.Match(expect)
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestDistance(t *testing.T) {
	state1 := StateOf("A", "B", "C")
	state2 := StateOf("A", "B", "D")
	state3 := StateOf("A", "B", "C", "D")
	state4 := StateOf("A", "B", "C")
	state5 := StateOf("A", "B", "C", "D", "E")

	assert.Equal(t, float32(100), state1.Distance(state2))
	assert.Equal(t, float32(100), state2.Distance(state1))
	assert.Equal(t, float32(100), state1.Distance(state3))
	assert.Equal(t, float32(0), state3.Distance(state1))
	assert.Equal(t, float32(0), state1.Distance(state4))
	assert.Equal(t, float32(0), state4.Distance(state1))
	assert.Equal(t, float32(200), state1.Distance(state5))
	assert.Equal(t, float32(0), state5.Distance(state1))
	assert.Equal(t, float32(200), state2.Distance(state5))
	assert.Equal(t, float32(0), state5.Distance(state2))
}

func TestStateString(t *testing.T) {
	state := StateOf("A", "B", "C")
	assert.Equal(t, "{C=100.00, B=100.00, A=100.00}", state.String())

	state = StateOf()
	assert.Equal(t, "{}", state.String())
}

func TestRemove(t *testing.T) {
	state := StateOf("A", "B", "C", "D", "E", "F", "G", "H", "I")

	state.Del("E")
	state.Del("F")
	assert.Equal(t, "{H=100.00, G=100.00, D=100.00, C=100.00, B=100.00, A=100.00, I=100.00}",
		state.String())
}
