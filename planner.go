// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root

package goap

import (
	"container/heap"
	"errors"
	"fmt"
)

// FindPlan finds a plan using the A* algorithm.
func FindPlan(start, goal State, actions []Action) (*Plan, error) {
	openSet := &nodeHeap{}
	heap.Init(openSet)
	startNode := &node{
		state:     start,
		cost:      0,
		heuristic: start.Distance(goal),
	}

	heap.Push(openSet, startNode)

	closedSet := make(map[uint64]struct{})

	for openSet.Len() > 0 {
		current := heap.Pop(openSet).(*node)

		// If we reached the goal, reconstruct the path.
		if current.state.HasAll(goal) {
			return reconstructPlan(current), nil
		}

		closedSet[current.state.Hash()] = struct{}{}

		for _, action := range actions {
			if !action.IsValid() { // TODO: remove from the list / filter out invalid actions at the beginning
				continue
			}

			if !current.state.HasAll(action.Require()) {
				continue
			}

			fmt.Println("current state", current.state.String())

			newState := current.state.Clone()
			newState.Apply(action.Outcome())

			fmt.Println("new state", newState.String())

			if _, found := closedSet[newState.Hash()]; found {
				continue
			}

			newCost := current.cost + action.Cost()

			// Check if newState is already in openSet or if the newCost is lower
			foundInOpenSet := false
			for _, openNode := range *openSet {
				if openNode.state.Equals(newState) {
					foundInOpenSet = true
					if newCost < openNode.cost {
						openNode.cost = newCost
						openNode.parent = current
						openNode.totalCost = newCost + openNode.heuristic
						heap.Fix(openSet, openNode.index) // Update the node's position in the heap
					}
					break
				}
			}

			if !foundInOpenSet {
				heuristic := newState.Distance(goal)
				newNode := &node{
					state:     newState,
					parent:    current,
					action:    action,
					cost:      newCost,
					heuristic: heuristic,
					totalCost: newCost + heuristic,
				}
				heap.Push(openSet, newNode)
			}
		}
	}

	return nil, errors.New("no plan could be found to reach the goal")
}

// reconstructPlan reconstructs the plan from the goal node to the start node.
func reconstructPlan(goalNode *node) *Plan {
	var actions []Action
	for n := goalNode; n != nil; n = n.parent {
		if n.action != nil { // The start node has no action
			actions = append(actions, n.action)
		}
	}
	// Reverse the slice of actions because we traversed the nodes from goal to start
	for i, j := 0, len(actions)-1; i < j; i, j = i+1, j-1 {
		actions[i], actions[j] = actions[j], actions[i]
	}
	return &Plan{actions: actions}
}

// ------------------------------------ Heap ------------------------------------

// node represents a node in the graph used by the A* algorithm.
type node struct {
	state     State
	action    Action  // The action that led to this node
	parent    *node   // Pointer to the parent node
	cost      float32 // Cost from the start node to this node
	heuristic float32 // Heuristic cost from this node to the goal
	totalCost float32 // Sum of cost and heuristic
	index     int     // Index of the node in the heap
}

// nodeHeap is a min-heap of nodes.
type nodeHeap []*node

func (h nodeHeap) Len() int           { return len(h) }
func (h nodeHeap) Less(i, j int) bool { return h[i].totalCost < h[j].totalCost }
func (h nodeHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i]; h[i].index = i; h[j].index = j }

func (h *nodeHeap) Push(x interface{}) {
	n := x.(*node)
	n.index = len(*h)
	*h = append(*h, n)
}

func (h *nodeHeap) Pop() interface{} {
	old := *h
	n := len(old)
	node := old[n-1]
	node.index = -1 // for safety
	*h = old[0 : n-1]
	return node
}
