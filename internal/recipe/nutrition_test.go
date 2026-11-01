package recipe_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lafeingcrokodil/diary/internal/food"
	"github.com/lafeingcrokodil/diary/internal/recipe"
)

func TestNutrition(t *testing.T) {
	t.Parallel()

	foods := map[string]food.Food{
		"lentils": {
			ID: "lentils", Unit: food.Grams, KcalPer100: 350,
			Nutrients: map[string]float64{"protein": 25, "fibre": 10},
		},
		"onion": {
			ID: "onion", Unit: food.Grams, KcalPer100: 40,
			Servings:  []food.Serving{{Name: "medium onion", Amount: 150}},
			Nutrients: map[string]float64{"protein": 1},
		},
		"stock": {ID: "stock", Unit: food.Millilitres, KcalPer100: 5},
	}

	tests := []struct {
		name        string
		ingredients []recipe.Ingredient
		want        recipe.Nutrition
	}{
		{
			name: "no ingredients",
			want: recipe.Nutrition{Nutrients: map[string]recipe.Total{}},
		},
		{
			name: "amounts in the food's unit and in servings",
			ingredients: []recipe.Ingredient{
				{FoodID: "lentils", Amount: 200},
				{FoodID: "onion", Serving: "medium onion", Amount: 2},
			},
			want: recipe.Nutrition{
				// 350 * 2 + 40 * 3
				Kcal: recipe.Total{Amount: 820},
				Nutrients: map[string]recipe.Total{
					// 25 * 2 + 1 * 3
					"protein": {Amount: 53},
					// The onion's fibre isn't known.
					"fibre": {Amount: 20, Missing: []int{1}},
				},
			},
		},
		{
			name: "the same food twice",
			ingredients: []recipe.Ingredient{
				{FoodID: "onion", Amount: 50},
				{FoodID: "lentils", Amount: 100},
				{FoodID: "onion", Serving: "medium onion", Amount: 1},
			},
			want: recipe.Nutrition{
				// 40 * 0.5 + 350 + 40 * 1.5
				Kcal: recipe.Total{Amount: 430},
				Nutrients: map[string]recipe.Total{
					"protein": {Amount: 27},
					"fibre":   {Amount: 10, Missing: []int{0, 2}},
				},
			},
		},
		{
			name:        "a food without nutrients",
			ingredients: []recipe.Ingredient{{FoodID: "stock", Amount: 1000}},
			want: recipe.Nutrition{
				Kcal:      recipe.Total{Amount: 50},
				Nutrients: map[string]recipe.Total{},
			},
		},
		{
			name: "a missing food and an unknown serving",
			ingredients: []recipe.Ingredient{
				{FoodID: "missing", Amount: 100},
				{FoodID: "lentils", Amount: 100},
				{FoodID: "onion", Serving: "small onion", Amount: 1},
			},
			want: recipe.Nutrition{
				Kcal: recipe.Total{Amount: 350, Missing: []int{0, 2}},
				Nutrients: map[string]recipe.Total{
					"protein": {Amount: 25, Missing: []int{0, 2}},
					"fibre":   {Amount: 10, Missing: []int{0, 2}},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			r := recipe.Recipe{Name: "Lentil soup", Ingredients: test.ingredients}
			got := r.Nutrition(foods)

			assert.InDelta(t, test.want.Kcal.Amount, got.Kcal.Amount, 1e-9)
			assert.Equal(t, test.want.Kcal.Missing, got.Kcal.Missing)
			assert.Len(t, got.Nutrients, len(test.want.Nutrients))

			for id, want := range test.want.Nutrients {
				assert.InDelta(t, want.Amount, got.Nutrients[id].Amount, 1e-9, id)
				assert.Equal(t, want.Missing, got.Nutrients[id].Missing, id)
			}
		})
	}
}
