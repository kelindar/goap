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

**GOAP**, standing for Goal-Oriented Action Planning, is a library developed in Go, designed to facilitate the creation of plans to achieve specific goals. Its primary application is in AI planning for areas such as game development, robotics, and other fields requiring complex decision-making. GOAP combines the simplicity of Go with the complexity of action planning, making it a practical tool for developers in need of reliable AI solutions.

### Key Features

1. **High Performance**: This library is built with performance in mind, ensuring it can handle real-time decision-making efficiently and with minimal overhead.

2. **Simple Interface**: This library is designed to be easy to use, with a simple interface that allows you to define actions and goals with minimal code.

3. **Supports Various State Types**: With support for both numeric and symbolic states, this library can be applied to a wide range of problems and use cases. This allows you to define states such as `food=10` or `!food` (no food).

4. **A\* Search**: Utilizing the A\* algorithm, the library efficiently navigates through possible plans, aiding in finding optimal paths to achieve goals. It uses the state distance heuristic to determine the best plan.

GOAP's blend of efficiency, flexibility, and practical search capabilities makes it a valuable tool for developers looking to incorporate advanced planning and decision-making into their AI applications.

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
