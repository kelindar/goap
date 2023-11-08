// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root

package goap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

/*
func TestSimplePlan(t *testing.T) {
	start := StateOf("A", "B")
	goal := StateOf("C", "D")
	actions := []Action{
		actionOf("Eat", 1.0, StateOf("hunger>50", "food>0"), StateOf("hunger-10", "food-1")),
		actionOf("Forage", 1.0, StateOf("food<10"), StateOf("tired+5", "food+1")),
		actionOf("Sleep", 1.0, StateOf("tired>50"), StateOf("tired-10")),

		// WONDER IF THIS CAN BE DONE INSIDE THE Predict() and Require() functions
	}

	plan, err := Plan(start, goal, actions)
	assert.NoError(t, err)
	assert.NotNil(t, plan)
	assert.Equal(t, 2, len(plan))
	assert.Equal(t, "Move A to D", plan[0].String())
	assert.Equal(t, "Move B to C", plan[1].String())
}*/

func TestSimplePlan(t *testing.T) {
	start := StateOf("A", "B")
	goal := StateOf("C", "D")
	actions := []Action{
		actionOf("Move A to C", 1.0, StateOf("A"), StateOf("!A", "C")),
		actionOf("Move A to D", 0.5, StateOf("A"), StateOf("!A", "D")),
		actionOf("Move B to C", 1.0, StateOf("B"), StateOf("!B", "C")),
		actionOf("Move B to D", 1.0, StateOf("B"), StateOf("!B", "D")),
	}

	plan, err := Plan(start, goal, actions)
	assert.NoError(t, err)
	assert.NotNil(t, plan)
	assert.Equal(t, 2, len(plan))
	assert.Equal(t, "Move A to D", plan[0].String())
	assert.Equal(t, "Move B to C", plan[1].String())
}

func TestNoPlanFound(t *testing.T) {
	start := StateOf("A", "B")
	goal := StateOf("C", "D")
	actions := []Action{
		actionOf("Move A to C", 1.0, StateOf("A"), StateOf("!A", "C")),
		actionOf("Move B to C", 1.0, StateOf("B"), StateOf("!B", "C")),
	}

	plan, err := Plan(start, goal, actions)
	assert.Error(t, err)
	assert.Nil(t, plan)
}

// ------------------------------------ Test Action ------------------------------------

func actionOf(name string, cost float32, require, outcome State) Action {
	return &testAction{
		name:    name,
		cost:    cost,
		require: require,
		outcome: outcome,
	}
}

type testAction struct {
	name    string
	cost    float32
	require State
	outcome State
}

func (a *testAction) Require() State {
	return a.require
}

func (a *testAction) Predict(_ State) State {
	return a.outcome
}

func (a *testAction) Cost() float32 {
	return a.cost
}

func (a *testAction) Perform() bool {
	return true
}

func (a *testAction) IsValid() bool {
	return true
}

func (a *testAction) String() string {
	return a.name
}
