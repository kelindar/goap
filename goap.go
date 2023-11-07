// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root

package goap

import "fmt"

type Goal interface {
	// IsAchieved checks if the goal has been achieved given the current state.
	IsAchieved(State) bool

	// Priority returns a numeric value representing the importance of the goal.
	// Higher values indicate more important goals.
	Priority() float32

	// DesiredState returns the State that represents the completion of the goal.
	DesiredState() State
}

type Action interface {
	fmt.Stringer
	Require() State
	Outcome() State
	IsValid() bool
	Perform() bool
	Cost() float32
}
