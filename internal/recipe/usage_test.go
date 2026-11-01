package recipe_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lafeingcrokodil/diary/internal/food"
	"github.com/lafeingcrokodil/diary/internal/recipe"
)

// usageRecipes returns recipes that use onions: one by weight, and one in
// medium onions. Soup also uses lentils, while bread doesn't use onions.
func usageRecipes() []recipe.Recipe {
	return []recipe.Recipe{
		{Name: "Bread", Ingredients: []recipe.Ingredient{{FoodID: "flour", Amount: 500}}},
		{Name: "Onion bhaji", Ingredients: []recipe.Ingredient{{FoodID: "onion", Amount: 300}}},
		{Name: "Lentil soup", Ingredients: []recipe.Ingredient{
			{FoodID: "lentils", Amount: 200},
			{FoodID: "onion", Serving: "medium onion", Amount: 1},
		}},
	}
}

func onion() food.Food {
	return food.Food{
		ID:       "onion",
		Name:     "Onion",
		Unit:     food.Grams,
		Servings: []food.Serving{{Name: "medium onion", Amount: 150}, {Name: "small onion", Amount: 80}},
	}
}

func TestCheckFoodUpdate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		recipes []recipe.Recipe
		change  func(f *food.Food)
		// wantErr is the error's message, or empty if there's no error.
		wantErr string
	}{
		{name: "no change", recipes: usageRecipes(), change: func(*food.Food) {}},
		{
			name:    "other fields",
			recipes: usageRecipes(),
			change: func(f *food.Food) {
				f.Name = "Red onion"
				f.KcalPer100 = 40
				f.Servings[0].Amount = 170
			},
		},
		{
			name:    "add a serving",
			recipes: usageRecipes(),
			change: func(f *food.Food) {
				f.Servings = append(f.Servings, food.Serving{Name: "large onion", Amount: 250})
			},
		},
		{
			name:    "remove a serving that no recipe uses",
			recipes: usageRecipes(),
			change:  func(f *food.Food) { f.Servings = f.Servings[:1] },
		},
		{
			name:    "reorder servings",
			recipes: usageRecipes(),
			change:  func(f *food.Food) { f.Servings[0], f.Servings[1] = f.Servings[1], f.Servings[0] },
		},
		{
			name:    "change the unit of a food that no recipe uses",
			recipes: usageRecipes()[:1],
			change:  func(f *food.Food) { f.Unit = food.Millilitres },
		},
		{
			name:    "change the unit",
			recipes: usageRecipes(),
			change:  func(f *food.Food) { f.Unit = food.Millilitres },
			wantErr: `food in use: the unit can't be changed, because the food is used in "Onion bhaji", "Lentil soup"`,
		},
		{
			name:    "rename a serving",
			recipes: usageRecipes(),
			change:  func(f *food.Food) { f.Servings[0].Name = "onion" },
			wantErr: `food in use: the serving "medium onion" can't be renamed or removed, because it's used in "Lentil soup"`,
		},
		{
			name:    "remove a serving",
			recipes: usageRecipes(),
			change:  func(f *food.Food) { f.Servings = f.Servings[1:] },
			wantErr: `food in use: the serving "medium onion" can't be renamed or removed, because it's used in "Lentil soup"`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			updated := onion()
			test.change(&updated)

			err := recipe.CheckFoodUpdate(test.recipes, onion(), updated)
			if test.wantErr == "" {
				assert.NoError(t, err)
				return
			}

			require.ErrorIs(t, err, recipe.ErrFoodInUse)
			assert.EqualError(t, err, test.wantErr)
		})
	}
}

func TestCheckFoodDeletion(t *testing.T) {
	t.Parallel()

	require.NoError(t, recipe.CheckFoodDeletion(usageRecipes(), "unused"))
	require.NoError(t, recipe.CheckFoodDeletion(nil, "onion"))

	err := recipe.CheckFoodDeletion(usageRecipes(), "onion")
	require.ErrorIs(t, err, recipe.ErrFoodInUse)
	assert.EqualError(t, err, `food in use: the food can't be deleted, because it's used in "Onion bhaji", "Lentil soup"`)
}
