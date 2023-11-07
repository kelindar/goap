// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root

package goap

import "errors"

// Plan holds a sequence of actions.
type Plan struct {
	actions []Action
}

// NewPlan creates a new Plan.
func NewPlan() *Plan {
	return &Plan{
		actions: make([]Action, 0),
	}
}

// AddAction adds an action to the end of the plan.
func (p *Plan) AddAction(action Action) {
	p.actions = append(p.actions, action)
}

// RemoveAction removes an action from the plan.
func (p *Plan) RemoveAction(action Action) error {
	for i, a := range p.actions {
		if a == action {
			p.actions = append(p.actions[:i], p.actions[i+1:]...)
			return nil
		}
	}
	return errors.New("action not found in plan")
}

// NextAction returns the next action to perform and removes it from the plan.
func (p *Plan) NextAction() (Action, error) {
	if len(p.actions) == 0 {
		return nil, errors.New("no more actions in the plan")
	}
	nextAction := p.actions[0]
	p.actions = p.actions[1:]
	return nextAction, nil
}

// HasActions returns true if there are actions left in the plan.
func (p *Plan) HasActions() bool {
	return len(p.actions) > 0
}

// IsValid checks if the plan is still valid given the current state.
func (p *Plan) IsValid(state State) bool {
	for _, action := range p.actions {
		preconditions := action.Require()
		for k, v := range preconditions {
			if stateValue, exists := state[k]; !exists || stateValue != v {
				return false
			}
		}
	}
	return true
}
