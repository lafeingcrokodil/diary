package server

import (
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/lafeingcrokodil/diary/internal/food"
	"github.com/lafeingcrokodil/diary/internal/recipe"
)

// recipeJSON is how the API represents a recipe. The owner is left out: it's
// always the signed-in user.
type recipeJSON struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	SourceURL   string           `json:"sourceUrl"`
	Ingredients []ingredientJSON `json:"ingredients"`
	Steps       []string         `json:"steps"`
	Notes       string           `json:"notes"`
	Archived    bool             `json:"archived"`
}

// ingredientJSON is an ingredient of a recipe. The amount is a number of
// servings if serving is set, and otherwise an amount in the food's unit.
type ingredientJSON struct {
	FoodID  string  `json:"foodId"`
	Serving string  `json:"serving"`
	Amount  float64 `json:"amount"`
	Note    string  `json:"note"`
}

// recipeResponseJSON is how the API responds with a recipe: as the user entered
// it, with its nutrition worked out from its foods' current values. Requests
// can't include the nutrition.
type recipeResponseJSON struct {
	recipeJSON

	Nutrition nutritionJSON `json:"nutrition"`
}

// nutritionJSON is the nutritional information of a whole recipe.
type nutritionJSON struct {
	Kcal totalJSON `json:"kcal"`
	// Nutrients maps nutrient IDs to their totals, for the nutrients whose
	// amount is known for at least one ingredient.
	Nutrients map[string]totalJSON `json:"nutrients"`
}

// totalJSON is a total over a recipe's ingredients. Missing holds the
// positions (from 0) of the ingredients whose amount isn't known, so that the
// amount is only a lower bound if there are any.
type totalJSON struct {
	Amount  float64 `json:"amount"`
	Missing []int   `json:"missing"`
}

func toTotalJSON(total recipe.Total) totalJSON {
	// An empty array rather than null, so that clients don't need to handle
	// both.
	missing := total.Missing
	if missing == nil {
		missing = []int{}
	}

	return totalJSON{Amount: total.Amount, Missing: missing}
}

// toRecipeResponse returns how the API responds with item, whose ingredients'
// foods are in foods, by ID.
func toRecipeResponse(item recipe.Recipe, foods map[string]food.Food) recipeResponseJSON {
	nutrition := item.Nutrition(foods)

	nutrients := make(map[string]totalJSON, len(nutrition.Nutrients))
	for id, total := range nutrition.Nutrients {
		nutrients[id] = toTotalJSON(total)
	}

	return recipeResponseJSON{
		recipeJSON: toRecipeJSON(item),
		Nutrition:  nutritionJSON{Kcal: toTotalJSON(nutrition.Kcal), Nutrients: nutrients},
	}
}

func toRecipeJSON(item recipe.Recipe) recipeJSON {
	ingredients := make([]ingredientJSON, 0, len(item.Ingredients))
	for _, ingredient := range item.Ingredients {
		ingredients = append(ingredients, ingredientJSON(ingredient))
	}

	// A recipe without steps has an empty array rather than null, so that
	// clients don't need to handle both.
	steps := item.Steps
	if steps == nil {
		steps = []string{}
	}

	return recipeJSON{
		ID:          item.ID,
		Name:        item.Name,
		SourceURL:   item.SourceURL,
		Ingredients: ingredients,
		Steps:       steps,
		Notes:       item.Notes,
		Archived:    item.Archived,
	}
}

func fromRecipeJSON(item recipeJSON) recipe.Recipe {
	ingredients := make([]recipe.Ingredient, 0, len(item.Ingredients))
	for _, ingredient := range item.Ingredients {
		ingredients = append(ingredients, recipe.Ingredient(ingredient))
	}

	return recipe.Recipe{
		ID:          item.ID,
		Name:        item.Name,
		SourceURL:   item.SourceURL,
		Ingredients: ingredients,
		Steps:       item.Steps,
		Notes:       item.Notes,
		Archived:    item.Archived,
	}
}

// handleRecipes adds the routes for the signed-in user's recipes to mux. A
// recipe's ingredients must be the user's own foods, which are kept in foods.
// Responses include each recipe's nutrition, worked out from its foods' current
// values. Recipes are archived and unarchived by updating them, and only
// archived recipes can be deleted (see recipe.Store.Delete).
func handleRecipes(mux *http.ServeMux, recipes recipe.Store, foods food.Store) {
	handlers := recipeHandlers{recipes: recipes, foods: foods}
	mux.HandleFunc("GET /api/recipes", handlers.list)
	mux.HandleFunc("POST /api/recipes", handlers.create)
	mux.HandleFunc("GET /api/recipes/{id}", handlers.get)
	mux.HandleFunc("PUT /api/recipes/{id}", handlers.update)
	mux.HandleFunc("DELETE /api/recipes/{id}", handlers.delete)
}

type recipeHandlers struct {
	recipes recipe.Store
	foods   food.Store
}

func (h recipeHandlers) list(resp http.ResponseWriter, req *http.Request) {
	list, err := h.recipes.List(req.Context(), owner(req))
	if err != nil {
		storeError(resp, req, err)
		return
	}

	foods, ok := h.foodsByID(resp, req)
	if !ok {
		return
	}

	result := make([]recipeResponseJSON, 0, len(list))
	for _, item := range list {
		result = append(result, toRecipeResponse(item, foods))
	}

	writeJSON(resp, req, http.StatusOK, result)
}

func (h recipeHandlers) create(resp http.ResponseWriter, req *http.Request) {
	input, foods, ok := h.readRecipe(resp, req)
	if !ok {
		return
	}

	created, err := h.recipes.Create(req.Context(), owner(req), input)
	if err != nil {
		storeError(resp, req, err)
		return
	}

	resp.Header().Set("Location", "/api/recipes/"+created.ID)
	writeJSON(resp, req, http.StatusCreated, toRecipeResponse(created, foods))
}

func (h recipeHandlers) get(resp http.ResponseWriter, req *http.Request) {
	got, err := h.recipes.Get(req.Context(), owner(req), req.PathValue("id"))
	if err != nil {
		storeError(resp, req, err)
		return
	}

	foods, ok := h.foodsByID(resp, req)
	if !ok {
		return
	}

	writeJSON(resp, req, http.StatusOK, toRecipeResponse(got, foods))
}

func (h recipeHandlers) update(resp http.ResponseWriter, req *http.Request) {
	input, foods, ok := h.readRecipe(resp, req)
	if !ok {
		return
	}

	// The path says which recipe to update, whatever the body says.
	input.ID = req.PathValue("id")

	updated, err := h.recipes.Update(req.Context(), owner(req), input)
	if err != nil {
		storeError(resp, req, err)
		return
	}

	writeJSON(resp, req, http.StatusOK, toRecipeResponse(updated, foods))
}

func (h recipeHandlers) delete(resp http.ResponseWriter, req *http.Request) {
	err := h.recipes.Delete(req.Context(), owner(req), req.PathValue("id"))
	if errors.Is(err, recipe.ErrNotArchived) {
		http.Error(resp, "Only archived recipes can be deleted.", http.StatusConflict)
		return
	}

	if err != nil {
		storeError(resp, req, err)
		return
	}

	resp.WriteHeader(http.StatusNoContent)
}

// foodsByID returns the signed-in user's foods, archived or not, by ID. If they
// can't be listed, it responds with a 500 and returns false.
func (h recipeHandlers) foodsByID(resp http.ResponseWriter, req *http.Request) (map[string]food.Food, bool) {
	list, err := h.foods.List(req.Context(), owner(req))
	if err != nil {
		storeError(resp, req, err)
		return nil, false
	}

	foods := make(map[string]food.Food, len(list))
	for _, item := range list {
		foods[item.ID] = item
	}

	return foods, true
}

// checkIngredients checks that a recipe's ingredients are all in foods, which
// holds the signed-in user's foods by ID, and that each named serving is one of
// its food's servings. If they aren't, it responds with a 400 and returns false.
func checkIngredients(resp http.ResponseWriter, item recipe.Recipe, foods map[string]food.Food) bool {
	for index, ingredient := range item.Ingredients {
		ingredientFood, ok := foods[ingredient.FoodID]
		if !ok {
			message := fmt.Sprintf("Ingredient %d: there's no food with the ID %q.", index+1, ingredient.FoodID)
			http.Error(resp, message, http.StatusBadRequest)

			return false
		}

		if ingredient.Serving == "" {
			continue
		}

		hasServing := slices.ContainsFunc(ingredientFood.Servings, func(serving food.Serving) bool {
			return serving.Name == ingredient.Serving
		})
		if !hasServing {
			message := fmt.Sprintf(
				"Ingredient %d: %s has no serving called %q.", index+1, ingredientFood.Name, ingredient.Serving,
			)
			http.Error(resp, message, http.StatusBadRequest)

			return false
		}
	}

	return true
}

// readRecipe decodes and validates the recipe in a request's body, and checks
// its ingredients against the signed-in user's foods, which it returns by ID.
// If that fails, it responds with a 400 (or a 500 if the foods can't be listed)
// and returns false.
func (h recipeHandlers) readRecipe(
	resp http.ResponseWriter, req *http.Request,
) (recipe.Recipe, map[string]food.Food, bool) {
	var input recipeJSON
	if !decodeBody(resp, req, &input, "recipe") {
		return recipe.Recipe{}, nil, false
	}

	parsed := fromRecipeJSON(input)
	if err := parsed.Validate(); err != nil {
		http.Error(resp, err.Error(), http.StatusBadRequest)
		return recipe.Recipe{}, nil, false
	}

	foods, ok := h.foodsByID(resp, req)
	if !ok || !checkIngredients(resp, parsed, foods) {
		return recipe.Recipe{}, nil, false
	}

	return parsed, foods, true
}
