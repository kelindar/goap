// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root

package goap

import (
	"errors"
	"sync"
)

// Action represents an action that can be performed.
type Action interface {

	// Simulate returns requirements and outcomes given the current
	// state (model) of the world.
	Simulate(current *State) (require, outcome *State)

	// Cost returns the cost of performing the action.
	Cost() float32
}

// Plan finds a plan to reach the goal from the start state using the provided actions.
func Plan(start, goal *State, actions []Action) ([]Action, error) {
	start = start.Clone()
	start.node = node{
		heuristic: start.Distance(goal),
		stateCost: 0,
	}

	heap := acquireHeap()
	heap.Push(start)
	defer heap.Release()

	for heap.Len() > 0 {
		current, _ := heap.Pop()

		// If we reached the goal, reconstruct the path.
		done, err := current.Match(goal)
		switch {
		case err != nil:
			return nil, err
		case done:
			return reconstructPlan(current), nil
		}

		for _, action := range actions {
			require, outcome := action.Simulate(current)
			match, err := current.Match(require)
			switch {
			case err != nil:
				return nil, err
			case !match:
				continue // Skip this action
			}

			// Apply the outcome to the new state
			newState := current.Clone()
			if err := newState.Apply(outcome); err != nil {
				return nil, err
			}

			//fmt.Printf("Action: %s, State: %s, New: %s\n", action.String(), current.String(), newState.String())

			// Check if newState is already planned to be visited or if the newCost is lower
			newCost := current.stateCost + action.Cost()
			node, found := heap.Find(newState.Hash())
			switch {
			case !found:
				heuristic := newState.Distance(goal)
				newState.parent = current
				newState.action = action
				newState.heuristic = heuristic
				newState.stateCost = newCost
				newState.totalCost = newCost + heuristic
				heap.Push(newState)

			// In any of those cases, we need to release the new state
			case found && !node.visited && newCost < node.stateCost:
				node.parent = current
				node.stateCost = newCost
				node.totalCost = newCost + node.heuristic
				heap.Fix(node) // Update the node's position in the heap
				fallthrough
			default: // The new state is already visited or the newCost is higher
				newState.release()
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

// ------------------------------------ Heap Pool ------------------------------------

var graphs = sync.Pool{
	New: func() any {
		return &graph{
			visit: make(map[uint32]*State, 32),
			heap:  make([]*State, 0, 32),
		}
	},
}

// Acquires a new instance of a heap
func acquireHeap() *graph {
	h := graphs.Get().(*graph)
	h.heap = h.heap[:0]
	clear(h.visit)
	return h
}

// Release the instance back to the pool
func (h *graph) Release() {
	for _, s := range h.visit {
		s.release()
	}
	graphs.Put(h)
}

// ------------------------------------ Heap ------------------------------------

type graph struct {
	visit map[uint32]*State
	heap  []*State
}

// Len returns the number of elements in the heap.
func (h *graph) Len() int { return len(h.heap) }

// Less reports whether the element with
func (h *graph) Less(i, j int) bool { return h.heap[i].totalCost < h.heap[j].totalCost }

// Swap swaps the elements with indexes i and j.
func (h *graph) Swap(i, j int) { h.heap[i], h.heap[j] = h.heap[j], h.heap[i] }

// Push pushes the element x onto the heap.
// The complexity is O(log n) where n = h.Len().
func (h *graph) Push(v *State) {
	v.index = h.Len()
	h.heap = append(h.heap, v)
	h.up(h.Len() - 1)
	h.visit[v.Hash()] = v
}

func (h *graph) Find(hash uint32) (*State, bool) {
	v, ok := h.visit[hash]
	return v, ok
}

// Pop removes and returns the minimum element (according to Less) from the heap.
// The complexity is O(log n) where n = h.Len().
// Pop is equivalent to Remove(h, 0).
func (h *graph) Pop() (*State, bool) {
	n := h.Len() - 1
	if n < 0 {
		return nil, false
	}

	h.Swap(0, n)
	h.down(0, n)
	return h.pop(), true
}

// Fix re-establishes the heap ordering after the element at index i has changed its value.
// Changing the value of the element at index i and then calling Fix is equivalent to,
// but less expensive than, calling Remove(h, i) followed by a Push of the new value.
// The complexity is O(log n) where n = h.Len().
func (h *graph) Fix(v *State) {
	if !h.down(v.index, h.Len()) {
		h.up(v.index)
	}
}

func (h *graph) pop() *State {
	old := h.heap
	n := len(old)
	node := old[n-1]
	node.visited = true

	h.heap = old[0 : n-1]
	h.visit[node.Hash()] = node
	return node
}

func (h *graph) up(j int) {
	for {
		i := (j - 1) / 2 // parent
		if i == j || !h.Less(j, i) {
			break
		}
		h.Swap(i, j)
		j = i
	}
}

func (h *graph) down(i0, n int) bool {
	i := i0
	for {
		j1 := 2*i + 1
		if j1 >= n || j1 < 0 { // j1 < 0 after int overflow
			break
		}
		j := j1 // left child
		if j2 := j1 + 1; j2 < n && h.Less(j2, j1) {
			j = j2 // = 2*i + 2  // right child
		}
		if !h.Less(j, i) {
			break
		}
		h.Swap(i, j)
		i = j
	}
	return i > i0
}
