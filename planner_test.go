// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root

package goap

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

/*
cpu: 13th Gen Intel(R) Core(TM) i7-13700K
BenchmarkPlan/deep-24         	     244	   4674232 ns/op	24156188 B/op	     117 allocs/op
BenchmarkPlan/deep-24         	    1750	    669815 ns/op	 5378442 B/op	      98 allocs/op
BenchmarkPlan/deep-24         	    2421	    475609 ns/op	 5378445 B/op	      98 allocs/op
BenchmarkPlan/deep-24         	    7272	    154309 ns/op	 1347781 B/op	      98 allocs/op
BenchmarkPlan/deep-24         	   26125	     45611 ns/op	  268996 B/op	     100 allocs/op
BenchmarkPlan/deep-24         	   34647	     34471 ns/op	    3898 B/op	      98 allocs/op
BenchmarkPlan/deep-24         	  199707	      5966 ns/op	    3763 B/op	      98 allocs/op
BenchmarkPlan/deep-24         	  255236	      4625 ns/op	    3285 B/op	      68 allocs/op
BenchmarkPlan/deep-24         	  251654	      4455 ns/op	    3205 B/op	      66 allocs/op
BenchmarkPlan/deep-24         	  284216	      4133 ns/op	    1426 B/op	      38 allocs/op
BenchmarkPlan/deep-24         	  311330	      3883 ns/op	    1426 B/op	      38 allocs/op
BenchmarkPlan/deep-24         	  339420	      3464 ns/op	     503 B/op	       5 allocs/op
BenchmarkPlan/deep-24         	  380756	      3103 ns/op	     230 B/op	       1 allocs/op

BenchmarkPlan/maze-24         	      37	  31458708 ns/op	 2702894 B/op	   80711 allocs/op
BenchmarkPlan/maze-24         	      63	  18643352 ns/op	 1569536 B/op	   51464 allocs/op
BenchmarkPlan/maze-24         	      64	  18393683 ns/op	 1628704 B/op	   51464 allocs/op
BenchmarkPlan/maze-24         	      69	  16841377 ns/op	  794077 B/op	   23827 allocs/op
*/
func BenchmarkPlan(b *testing.B) {
	b.ReportAllocs()

	b.Run("deep", func(b *testing.B) {
		start := StateOf("hunger=80", "!food", "!tired")
		goal := StateOf("food>80")
		actions := []Action{
			actionOf("Eat", 1.0, StateOf("food>0"), StateOf("hunger-50", "food-5")),
			actionOf("Forage", 1.0, StateOf("tired<50"), StateOf("tired+20", "food+10", "hunger+5")),
			actionOf("Sleep", 1.0, StateOf("tired>30"), StateOf("tired-50")),
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := Plan(start, goal, actions)
			assert.NoError(b, err)
		}
	})

	b.Run("maze", func(b *testing.B) {
		start := StateOf("A")
		goal := StateOf("Z")
		actions := []Action{
			move("A->B"), move("B->C"), move("C->D"), move("D->E"), move("E->F"), move("F->G"),
			move("G->H"), move("H->I"), move("I->J"), move("C->X1"), move("E->X2"), move("G->X3"),
			move("X1->D"), move("X2->F"), move("X3->H"), move("B->Y1"), move("D->Y2"), move("F->Y3"),
			move("Y1->C"), move("Y2->E"), move("Y3->G"), move("J->K"), move("K->L"), move("L->M"),
			move("M->N"), move("N->O"), move("O->P"), move("P->Q"), move("Q->R"), move("R->S"),
			move("S->T"), move("T->U"), move("U->V"), move("V->W"), move("W->X"), move("X->Y"),
			move("Y->Z"), move("U->Z1"), move("W->Z2"), move("Z1->V"), move("Z2->X"), move("A->Z3"),
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := Plan(start, goal, actions)
			assert.NoError(b, err)
		}
	})
}

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

func TestMaze(t *testing.T) {
	start := StateOf("A")
	goal := StateOf("Z")
	actions := []Action{
		move("A->B"), move("B->C"), move("C->D"), move("D->E"), move("E->F"), move("F->G"),
		move("G->H"), move("H->I"), move("I->J"), move("C->X1"), move("E->X2"), move("G->X3"),
		move("X1->D"), move("X2->F"), move("X3->H"), move("B->Y1"), move("D->Y2"), move("F->Y3"),
		move("Y1->C"), move("Y2->E"), move("Y3->G"), move("J->K"), move("K->L"), move("L->M"),
		move("M->N"), move("N->O"), move("O->P"), move("P->Q"), move("Q->R"), move("R->S"),
		move("S->T"), move("T->U"), move("U->V"), move("V->W"), move("W->X"), move("X->Y"),
		move("Y->Z"), move("U->Z1"), move("W->Z2"), move("Z1->V"), move("Z2->X"), move("A->Z3"),
	}

	plan, err := Plan(start, goal, actions)
	assert.NoError(t, err)
	assert.Equal(t, []string{"A->B", "B->C", "C->D", "D->E", "E->F", "F->G", "G->H", "H->I", "I->J",
		"J->K", "K->L", "L->M", "M->N", "N->O", "O->P", "P->Q", "Q->R", "R->S", "S->T", "T->U", "U->V",
		"V->W", "W->X", "X->Y", "Y->Z"},
		planOf(plan))
}

func TestSimplePlan(t *testing.T) {
	start := StateOf("A", "B")
	goal := StateOf("C", "D")
	actions := []Action{move("A->C"), move("A->D"), move("B->C"), move("B->D")}

	plan, err := Plan(start, goal, actions)
	assert.NoError(t, err)
	assert.Equal(t, []string{"A->C", "B->D"},
		planOf(plan))
}

func TestNoPlanFound(t *testing.T) {
	plan, err := Plan(StateOf("A", "B"), StateOf("C", "D"), []Action{
		move("A->C"), move("B->C"),
	})
	assert.Error(t, err)
	assert.Nil(t, plan)
}

// ------------------------------------ Test Action ------------------------------------

func move(m string) Action {
	arr := strings.Split(m, "->")
	return actionOf(m, 1, StateOf(arr[0]), StateOf("!"+arr[0], arr[1]))
}

func planOf(plan []Action) []string {
	var result []string
	for _, action := range plan {
		result = append(result, action.String())
	}
	return result
}

func actionOf(name string, cost float32, require, outcome *State) Action {
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
	require *State
	outcome *State
}

func (a *testAction) Simulate(_ *State) (*State, *State) {
	return a.require, a.outcome
}

func (a *testAction) Cost() float32 {
	return a.cost
}

func (a *testAction) String() string {
	return a.name
}
