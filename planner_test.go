// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root

package goap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindPlan(t *testing.T) {
	start := StateOf("A", "B")
	goal := StateOf("C", "D")
	actions := []Action{
		actionOf("Move A to C", 1.0, StateOf("A"), StateOf("C")),
		actionOf("Move A to D", 1.0, StateOf("A"), StateOf("D")),
		actionOf("Move B to C", 1.0, StateOf("B"), StateOf("C")),
		actionOf("Move B to D", 1.0, StateOf("B"), StateOf("D")),
	}

	plan, err := FindPlan(start, goal, actions)
	assert.NoError(t, err)
	assert.NotNil(t, plan)
	assert.Equal(t, 2, len(plan.actions))
	assert.Equal(t, "Move A to C", plan.actions[0].String())
	assert.Equal(t, "Move B to D", plan.actions[1].String())
}

func TestNoPlanFound(t *testing.T) {
	start := StateOf("A", "B")
	goal := StateOf("C", "D")
	actions := []Action{
		actionOf("Move A to C", 1.0, StateOf("A"), StateOf("C")),
		actionOf("Move B to C", 1.0, StateOf("B"), StateOf("C")),
	}

	plan, err := FindPlan(start, goal, actions)
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

func (a *testAction) Outcome() State {
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
