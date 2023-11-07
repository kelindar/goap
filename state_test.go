// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root

package goap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStateOf(t *testing.T) {
	state := StateOf("A", "B", "C")
	assert.Equal(t, 3, len(state))
	assert.True(t, state.Has("A"))
	assert.True(t, state.Has("B"))
	assert.True(t, state.Has("C"))
	assert.False(t, state.Has("D"))
}

func TestHasAll(t *testing.T) {
	state1 := StateOf("A", "B", "C")
	state2 := StateOf("A", "B")
	state3 := StateOf("A", "B", "C", "D")

	assert.True(t, state1.HasAll(state2))
	assert.False(t, state2.HasAll(state1))
	assert.True(t, state3.HasAll(state1))
	assert.False(t, state1.HasAll(state3))
	assert.True(t, state3.HasAll(state2))
	assert.False(t, state2.HasAll(state3))
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

	assert.True(t, state1.Has("A"))
	assert.True(t, state1.Has("B"))
	assert.True(t, state1.Has("C"))
	assert.True(t, state1.Has("D"))
	assert.True(t, state1.Has("E"))
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

func TestStateHash(t *testing.T) {
	state1 := StateOf("A", "B", "C")
	state2 := StateOf("C", "B", "A")
	state3 := StateOf("A", "B", "C", "D")

	assert.Equal(t, state1.Hash(), state2.Hash())
	assert.NotEqual(t, state1.Hash(), state3.Hash())
	assert.NotEqual(t, state2.Hash(), state3.Hash())
}
