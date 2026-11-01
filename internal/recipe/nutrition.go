package recipe

import (
	"slices"

	"github.com/lafeingcrokodil/diary/internal/food"
)

// Total is the total of an amount (such as energy, or a nutrient) over a
// recipe's ingredients.
type Total struct {
	// Amount is the sum over the ingredients whose amount is known.
	Amount float64
	// Missing holds the positions (from 0) of the ingredients whose amount
	// isn't known, so that Amount is only a lower bound if there are any.
	Missing []int
}

// Nutrition is the nutritional information of a whole recipe.
type Nutrition struct {
	Kcal Total
	// Nutrients maps nutrient IDs to their totals, in each nutrient's own
	// unit. It only has the nutrients whose amount is known for at least one
	// ingredient.
	Nutrients map[string]Total
}

// Nutrition works out the nutritional information of the whole recipe from the
// current values of its foods, which foods holds by ID. An ingredient whose
// food or serving isn't in foods counts as missing from every total, rather
// than as none, since how much it contributes isn't known.
func (r Recipe) Nutrition(foods map[string]food.Food) Nutrition {
	ingredients := r.resolve(foods)

	nutrition := Nutrition{
		Kcal:      total(ingredients, func(item food.Food) (float64, bool) { return item.KcalPer100, true }),
		Nutrients: map[string]Total{},
	}

	for _, ingredient := range ingredients {
		if ingredient.food == nil {
			continue
		}

		for id := range ingredient.food.Nutrients {
			if _, done := nutrition.Nutrients[id]; done {
				continue
			}

			nutrition.Nutrients[id] = total(ingredients, func(item food.Food) (float64, bool) {
				amount, known := item.Nutrients[id]
				return amount, known
			})
		}
	}

	return nutrition
}

// resolvedIngredient is an ingredient with its food, or nil if the food or
// serving can't be found, and its amount in the food's unit.
type resolvedIngredient struct {
	food   *food.Food
	amount float64
}

// resolve looks up the foods of the recipe's ingredients in foods, which holds
// them by ID.
func (r Recipe) resolve(foods map[string]food.Food) []resolvedIngredient {
	resolved := make([]resolvedIngredient, 0, len(r.Ingredients))

	for _, ingredient := range r.Ingredients {
		item, found := foods[ingredient.FoodID]
		if !found {
			resolved = append(resolved, resolvedIngredient{})
			continue
		}

		amount, found := ingredient.amountIn(item)
		if !found {
			resolved = append(resolved, resolvedIngredient{})
			continue
		}

		resolved = append(resolved, resolvedIngredient{food: &item, amount: amount})
	}

	return resolved
}

// total adds up an amount over ingredients, where valuePer100 returns the
// amount per 100 of a food's unit, and false if it isn't known.
func total(ingredients []resolvedIngredient, valuePer100 func(food.Food) (float64, bool)) Total {
	var result Total

	for index, ingredient := range ingredients {
		value, known := 0.0, false
		if ingredient.food != nil {
			value, known = valuePer100(*ingredient.food)
		}

		if known {
			result.Amount += value * ingredient.amount / per100
		} else {
			result.Missing = append(result.Missing, index)
		}
	}

	return result
}

// per100 is the amount of a food that its nutritional values are given for.
const per100 = 100

// amountIn returns the ingredient's amount in item's unit, and false if the
// ingredient's serving isn't one of item's servings.
func (i Ingredient) amountIn(item food.Food) (float64, bool) {
	if i.Serving == "" {
		return i.Amount, true
	}

	index := slices.IndexFunc(item.Servings, func(serving food.Serving) bool {
		return serving.Name == i.Serving
	})
	if index < 0 {
		return 0, false
	}

	return i.Amount * item.Servings[index].Amount, true
}
