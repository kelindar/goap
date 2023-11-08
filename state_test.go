// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root

package goap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

/*
cpu: 13th Gen Intel(R) Core(TM) i7-13700K
BenchmarkState/clone-24         	11981906	       101.0 ns/op	     144 B/op	       2 allocs/op
BenchmarkState/has-24           	39235688	        30.17 ns/op	       0 B/op	       0 allocs/op
BenchmarkState/add-24           	 6542553	       187.3 ns/op	     160 B/op	       6 allocs/op
BenchmarkState/remove-24        	 6646038	       178.7 ns/op	     160 B/op	       6 allocs/op
BenchmarkState/apply-24         	28536843	        40.25 ns/op	       0 B/op	       0 allocs/op
BenchmarkState/distance-24      	29706205	        40.97 ns/op	       0 B/op	       0 allocs/op
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

	b.Run("has", func(b *testing.B) {
		state1 := StateOf("A", "B", "C")
		state2 := StateOf("A", "B")
		for i := 0; i < b.N; i++ {
			state1.Has(state2)
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
			state.Remove("A")
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

func TestHas(t *testing.T) {
	state1 := StateOf("A", "B", "C")
	state2 := StateOf("A", "B")
	state3 := StateOf("A", "B", "C", "D")

	assert.True(t, state1.Has(state2))
	assert.False(t, state2.Has(state1))
	assert.True(t, state3.Has(state1))
	assert.False(t, state1.Has(state3))
	assert.True(t, state3.Has(state2))
	assert.False(t, state2.Has(state3))
}

func TestStateHash(t *testing.T) {
	state1 := StateOf("A", "B", "C")
	state2 := StateOf("C", "B", "A")
	state3 := StateOf("A", "B", "C", "D")

	assert.Equal(t, state1.Hash(), state2.Hash())
	assert.NotEqual(t, state1.Hash(), state3.Hash())
	assert.NotEqual(t, state2.Hash(), state3.Hash())
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
	state.Remove("A")
	assert.False(t, clone.Equals(state))
}
func TestStateApply(t *testing.T) {
	state1 := StateOf("A", "B", "C")
	state2 := StateOf("D", "E")
	state1.Apply(state2)

	expect := StateOf("A", "B", "C", "D", "E")
	assert.True(t, state1.Has(expect))
}

func TestDistance(t *testing.T) {
	state1 := StateOf("A", "B", "C")
	state2 := StateOf("A", "B", "D")
	state3 := StateOf("A", "B", "C", "D")
	state4 := StateOf("A", "B", "C")
	state5 := StateOf("A", "B", "C", "D", "E")

	assert.Equal(t, float32(1), state1.Distance(state2))
	assert.Equal(t, float32(1), state2.Distance(state1))
	assert.Equal(t, float32(1), state1.Distance(state3))
	assert.Equal(t, float32(0), state3.Distance(state1))
	assert.Equal(t, float32(0), state1.Distance(state4))
	assert.Equal(t, float32(0), state4.Distance(state1))
	assert.Equal(t, float32(2), state1.Distance(state5))
	assert.Equal(t, float32(0), state5.Distance(state1))
	assert.Equal(t, float32(2), state2.Distance(state5))
	assert.Equal(t, float32(0), state5.Distance(state2))
}
