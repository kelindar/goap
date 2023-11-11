<p align="center">
<img width="270" height="110" src=".github/logo.png" border="0" alt="kelindar/goap">
<br>
<img src="https://img.shields.io/github/go-mod/go-version/kelindar/goap" alt="Go Version">
<a href="https://pkg.go.dev/github.com/kelindar/goap"><img src="https://pkg.go.dev/badge/github.com/kelindar/goap" alt="PkgGoDev"></a>
<a href="https://goreportcard.com/report/github.com/kelindar/goap"><img src="https://goreportcard.com/badge/github.com/kelindar/goap" alt="Go Report Card"></a>
<a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License"></a>
<a href="https://coveralls.io/github/kelindar/goap"><img src="https://coveralls.io/repos/github/kelindar/goap/badge.svg" alt="Coverage"></a>
</p>

## GOAP: Goal-Oriented Action Planning in Go

**GOAP** is a Goal-Oriented Action Planning library written in Go. It is designed to help you find plans to achieve a specific goal by selecting appropriate actions from a set of available actions. GOAP is well-suited for solving complex decision-making problems, such as AI planning in games, robotics, and other applications.

### Features

1. **High Performance**: This library is designed for efficiency and performance, making it suitable for real-time applications.

2. **Flexible Actions**: Define custom actions that can be used in the planning process.

3. **Numeric and Symbolic States**: The library supports both numeric and symbolic states, making it versatile for various types of problems

4. **Heuristic Search**: The library uses heuristic search (A\*) to efficiently explore possible plans.

## Quick Start

This guide will walk you through implementing a simple AI using the GOAP (Goal-Oriented Action Planning) library. We'll create a scenario where an AI character must manage hunger and tiredness while gathering food. Our AI character starts with high hunger and no food. The goal is to accumulate food (`food>80`) while managing hunger and tiredness.

### Step 1: Initialize States

Start by defining the initial state (`init`) and the goal (`goal`):

```go
init := goap.StateOf("hunger=80", "!food", "!tired")
goal := goap.StateOf("food>80")
```

- `init` represents the AI's starting condition: hunger at 80, no food, and not tired.
- `goal` is the desired state where the AI has more than 80 units of food.

### Step 2: Define Actions

Define the actions the AI can perform: `eat`, `forage`, and `sleep`. Each action has prerequisites and outcomes. First, we need to implement the `Action` interface. For simplicity, we create a generic action:

```go
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
```

Next, we can define a list of actions that the AI can perform (`eat`, `forage`, and `sleep`).

```go
actions := []goap.Action{
    NewAction("eat", "food>0", "hunger-50,food-5"),
    NewAction("forage", "tired<50", "tired+20,food+10,hunger+5"),
    NewAction("sleep", "tired>30", "tired-30"),
}
```

- `eat`: Can be performed if `food>0`. Reduces hunger by 50 and food by 5.
- `forage`: Possible when `tired<50`. Increases tiredness by 20, food by 10, and hunger by 5.
- `sleep`: Executable if `tired>30`. Reduces tiredness by 30.

### Step 3: Generate the Plan

Generate a plan to reach the goal from the initial state using the defined actions.

```go
plan, err := goap.Plan(init, goal, actions)
if err != nil {
    panic(err)
}
```

The output will be a sequence of actions that the AI character needs to perform to achieve its goal. Here's an example:

```text
1. forage
2. forage
3. forage
4. sleep
5. forage
6. sleep
7. forage
8. forage
9. sleep
10. forage
11. sleep
12. forage
13. eat
14. forage
```

This sequence represents the AI's decision-making process, balancing foraging for food, eating to reduce hunger, and sleeping to manage tiredness.

## License

This library is licensed under the MIT license. See the [LICENSE](https://github.com/kelindar/goap/LICENSE) file in the project root for more details.

## Credits

This library is developed and maintained by [Roman Atachiants](https://github.com/kelindar) and contributions from the open-source community. We welcome contributions and feedback to make the library even better.
