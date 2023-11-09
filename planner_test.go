// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root

package goap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNumericPlan(t *testing.T) {
	start := StateOf("hunger=80", "!food", "!tired")
	goal := StateOf("food>80")
	actions := []Action{
		actionOf("Eat", 1.0, StateOf("food>0"), StateOf("hunger-50", "food-5")),
		actionOf("Forage", 1.0, StateOf("tired<50"), StateOf("tired+20", "food+10", "hunger+5")),
		actionOf("Sleep", 1.0, StateOf("tired>30"), StateOf("tired-50")),
	}

	plan, err := Plan(start, goal, actions)
	assert.NoError(t, err)
	assert.Equal(t, []string{"Forage", "Forage", "Forage", "Sleep", "Forage", "Forage", "Sleep", "Forage", "Forage", "Forage", "Sleep", "Eat", "Forage"},
		planOf(plan))
}

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
	assert.Equal(t, []string{"Move A to D", "Move B to C"},
		planOf(plan))
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

func planOf(plan []Action) []string {
	var result []string
	for _, action := range plan {
		result = append(result, action.String())
	}
	return result
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
