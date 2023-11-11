package main

import (
	"fmt"
	"strings"

	"github.com/kelindar/goap"
)

func main() {
	init := goap.StateOf("hunger=80", "!food", "!tired")
	goal := goap.StateOf("food>80")

	actions := []goap.Action{
		NewAction("eat", "food>0", "hunger-50,food-5"),
		NewAction("forage", "tired<50", "tired+20,food+10,hunger+5"),
		NewAction("sleep", "tired>30", "tired-30"),
	}

	plan, err := goap.Plan(init, goal, actions)
	if err != nil {
		panic(err)
	}

	for i, action := range plan {
		fmt.Printf("%2d. %s\n", i+1, action.(*Action).String())
	}
}

// NewAction creates a new action from the given name, require and outcome.
func NewAction(name, require, outcome string) *Action {
	return &Action{
		name:    name,
		require: goap.StateOf(strings.Split(require, ",")...),
		outcome: goap.StateOf(strings.Split(outcome, ",")...),
	}
}

// Action represents a single action that can be performed by the agent.
type Action struct {
	name    string
	cost    int
	require *goap.State
	outcome *goap.State
}

// Simulate simulates the action and returns the required and outcome states.
func (a *Action) Simulate(current *goap.State) (*goap.State, *goap.State) {
	return a.require, a.outcome
}

// Cost returns the cost of the action.
func (a *Action) Cost() float32 {
	return 1
}

// String returns the name of the action.
func (a *Action) String() string {
	return a.name
}
