// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root

package goap

import (
	"container/heap"
	"errors"
	"fmt"
)

// Action represents an action that can be performed.
type Action interface {
	fmt.Stringer

	// Simulate returns requirements and outcomes given the current
	// state (model) of the world.
	Simulate(current *State) (require, outcome *State)

	// Cost returns the cost of performing the action.
	Cost() float32
}

// Plan finds a plan to reach the goal from the start state using the provided actions.
func Plan(start, goal *State, actions []Action) ([]Action, error) {
	visited := make(map[uint32]struct{}, 32)
	pending := &stateHeap{}
	heap.Init(pending)
	heap.Push(pending, &State{
		hx: start.hx,
		vx: start.vx,
		node: node{
			stateCost: 0,
			heuristic: start.Distance(goal),
		},
	})

	for pending.Len() > 0 {
		current := heap.Pop(pending).(*State)

		// If we reached the goal, reconstruct the path.
		done, err := current.Match(goal)
		switch {
		case err != nil:
			return nil, err
		case done:
			return reconstructPlan(current), nil
		}

		visited[current.Hash()] = struct{}{}

		for _, action := range actions {
			require, outcome := action.Simulate(current)

			// fmt.Printf("Explore %s\n", action.String())

			// Check if the current state satisfies the action's requirements
			match, err := current.Match(require)
			switch {
			case err != nil:
				return nil, err
			case !match:
				continue // Skip this action
			}

			newState := current.Clone()
			defer newState.release()

			// Apply the outcome to the new state
			if err := newState.Apply(outcome); err != nil {
				return nil, err
			}

			//fmt.Printf("Action: %s, State: %s, New: %s\n", action.String(), current.String(), newState.String())

			// If the new state was already visited, skip it
			if _, ok := visited[newState.Hash()]; ok {
				continue
			}

			newCost := current.stateCost + action.Cost()

			// Check if newState is already planned to be visited or if the newCost is lower
			foundInPending := false
			for _, node := range *pending {
				if node.Equals(newState) {
					foundInPending = true
					if newCost < node.stateCost {
						node.parent = current
						node.stateCost = newCost
						node.totalCost = newCost + node.heuristic
						pending.Update(node) // Update the node's position in the heap
					}
					break
				}
			}

			if !foundInPending {
				heuristic := newState.Distance(goal)
				newState.parent = current
				newState.action = action
				newState.heuristic = heuristic
				newState.stateCost = newCost
				newState.totalCost = newCost + heuristic
				heap.Push(pending, newState)
			}
		}
	}

	return nil, errors.New("no plan could be found to reach the goal")
}

// reconstructPlan reconstructs the plan from the goal node to the start node.
func reconstructPlan(goalNode *State) []Action {
	plan := make([]Action, 0, int(goalNode.index))
	for n := goalNode; n != nil; n = n.parent {
		if n.action != nil { // The start node has no action
			plan = append(plan, n.action)
		}
	}

	// Reverse the slice of actions because we traversed the nodes from goal to start
	for i, j := 0, len(plan)-1; i < j; i, j = i+1, j-1 {
		plan[i], plan[j] = plan[j], plan[i]
	}
	return plan
}

// ------------------------------------ Heap ------------------------------------

// stateHeap is a min-heap of states.
type stateHeap []*State

func (h stateHeap) Len() int           { return len(h) }
func (h stateHeap) Less(i, j int) bool { return h[i].totalCost < h[j].totalCost }
func (h stateHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i]; h[i].index = i; h[j].index = j }

func (h *stateHeap) Push(x any) {
	n := x.(*State)
	n.index = len(*h)
	*h = append(*h, n)
}

func (h *stateHeap) Pop() any {
	old := *h
	n := len(old)
	node := old[n-1]
	*h = old[0 : n-1]
	return node
}

// Update modifies the priority and value of an element in the queue.
func (h *stateHeap) Update(n *State) {
	heap.Fix(h, n.index)
}
