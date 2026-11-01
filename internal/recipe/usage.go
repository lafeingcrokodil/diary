package recipe

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/lafeingcrokodil/diary/internal/food"
)

// ErrFoodInUse is wrapped by the errors that CheckFoodUpdate and
// CheckFoodDeletion return.
var ErrFoodInUse = errors.New("food in use")

// CheckFoodUpdate checks that changing a food from old to updated keeps the
// meaning of the amounts in recipes, which must be all of the food's owner's
// recipes, archived or not. Recipes' amounts are in the food's unit or its
// servings, so a food that a recipe uses can't change its unit, and a serving
// that a recipe uses can't be renamed or removed. If the change isn't allowed,
// it returns an error that wraps ErrFoodInUse and names the recipes.
func CheckFoodUpdate(recipes []Recipe, old, updated food.Food) error {
	if updated.Unit != old.Unit {
		if users := usersOf(recipes, old.ID, anyServing); len(users) > 0 {
			return fmt.Errorf("%w: the unit can't be changed, because the food is used in %s", ErrFoodInUse, users)
		}
	}

	for _, serving := range old.Servings {
		kept := slices.ContainsFunc(updated.Servings, func(s food.Serving) bool { return s.Name == serving.Name })
		if kept {
			continue
		}

		inServing := func(ingredient Ingredient) bool { return ingredient.Serving == serving.Name }
		if users := usersOf(recipes, old.ID, inServing); len(users) > 0 {
			return fmt.Errorf(
				"%w: the serving %q can't be renamed or removed, because it's used in %s",
				ErrFoodInUse, serving.Name, users,
			)
		}
	}

	return nil
}

// CheckFoodDeletion checks that none of recipes, which must be all of the
// food's owner's recipes, archived or not, uses the food with the given ID. If
// any does, it returns an error that wraps ErrFoodInUse and names them.
func CheckFoodDeletion(recipes []Recipe, foodID string) error {
	if users := usersOf(recipes, foodID, anyServing); len(users) > 0 {
		return fmt.Errorf("%w: the food can't be deleted, because it's used in %s", ErrFoodInUse, users)
	}

	return nil
}

func anyServing(Ingredient) bool { return true }

// recipeNames are the names of some recipes, which print as a list.
type recipeNames []string

func (n recipeNames) String() string {
	quoted := make([]string, 0, len(n))
	for _, name := range n {
		quoted = append(quoted, strconv.Quote(name))
	}

	return strings.Join(quoted, ", ")
}

// usersOf returns the names of the recipes with an ingredient of the food with
// the given ID for which matches returns true.
func usersOf(recipes []Recipe, foodID string, matches func(Ingredient) bool) recipeNames {
	var names recipeNames

	for _, r := range recipes {
		uses := slices.ContainsFunc(r.Ingredients, func(ingredient Ingredient) bool {
			return ingredient.FoodID == foodID && matches(ingredient)
		})
		if uses {
			names = append(names, r.Name)
		}
	}

	return names
}
