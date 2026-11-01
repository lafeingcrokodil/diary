package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lafeingcrokodil/diary/internal/food/foodtest"
	"github.com/lafeingcrokodil/diary/internal/nutrient/nutrienttest"
	"github.com/lafeingcrokodil/diary/internal/recipe"
	"github.com/lafeingcrokodil/diary/internal/recipe/recipetest"
	"github.com/lafeingcrokodil/diary/internal/server"
)

type recipeJSON struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	SourceURL   string `json:"sourceUrl"`
	Ingredients []struct {
		FoodID  string  `json:"foodId"`
		Serving string  `json:"serving"`
		Amount  float64 `json:"amount"`
		Note    string  `json:"note"`
	} `json:"ingredients"`
	Steps     []string      `json:"steps"`
	Notes     string        `json:"notes"`
	Archived  bool          `json:"archived"`
	Nutrition nutritionJSON `json:"nutrition"`
}

type nutritionJSON struct {
	Kcal      totalJSON            `json:"kcal"`
	Nutrients map[string]totalJSON `json:"nutrients"`
}

type totalJSON struct {
	Amount  float64 `json:"amount"`
	Missing []int   `json:"missing"`
}

// newRecipesAPI returns an api whose recipes are kept in store, and whose foods
// and nutrients are kept in memory.
func newRecipesAPI(store recipe.Store) api {
	return newAPI(server.Stores{
		Foods:     foodtest.NewMemoryStore(),
		Nutrients: nutrienttest.NewMemoryStore(),
		Recipes:   store,
	})
}

// sandwich returns a recipe made with two slices of the food with the given ID
// (bread, which has a "slice" serving) and 10 of its unit.
func sandwich(breadID string) string {
	return `{"name":"Open sandwich","sourceUrl":"https://example.com/sandwich",` +
		`"ingredients":[{"foodId":"` + breadID + `","serving":"slice","amount":2,"note":"toasted"},` +
		`{"foodId":"` + breadID + `","serving":"","amount":10,"note":"as crumbs"}],` +
		`"steps":["Toast the bread.","Sprinkle the crumbs on top."],"notes":"Best fresh.","archived":false}`
}

// createSandwich adds bread as one of alice's foods, and a sandwich made with
// it as one of her recipes, and returns the recipe.
func (a api) createSandwich(t *testing.T) recipeJSON {
	t.Helper()

	breadID := a.createBread(t).ID

	rec := a.do(t, http.MethodPost, "/api/recipes", alice, sandwich(breadID))
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	return decodeRecipe(t, rec)
}

func decodeRecipe(t *testing.T, rec *httptest.ResponseRecorder) recipeJSON {
	t.Helper()

	var result recipeJSON
	decode(t, rec, &result)

	return result
}

func TestRecipes(t *testing.T) {
	t.Parallel()

	t.Run("create, get and list", func(t *testing.T) {
		t.Parallel()

		recipes := newRecipesAPI(recipetest.NewMemoryStore())
		breadID := recipes.createBread(t).ID

		rec := recipes.do(t, http.MethodPost, "/api/recipes", alice, sandwich(breadID))
		require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

		created := decodeRecipe(t, rec)
		assert.NotEmpty(t, created.ID)
		assert.Equal(t, "/api/recipes/"+created.ID, rec.Header().Get("Location"))
		assert.JSONEq(t, sandwich(breadID), withoutFields(t, rec.Body.String(), "id", "nutrition"))

		rec = recipes.do(t, http.MethodGet, "/api/recipes/"+created.ID, alice, "")
		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, created, decodeRecipe(t, rec))

		rec = recipes.do(t, http.MethodGet, "/api/recipes", alice, "")
		require.Equal(t, http.StatusOK, rec.Code)

		var list []recipeJSON
		decode(t, rec, &list)
		assert.Equal(t, []recipeJSON{created}, list)
	})

	t.Run("list without recipes", func(t *testing.T) {
		t.Parallel()

		rec := newRecipesAPI(recipetest.NewMemoryStore()).do(t, http.MethodGet, "/api/recipes", alice, "")
		require.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, "[]", rec.Body.String())
	})

	t.Run("recipe without ingredients or steps", func(t *testing.T) {
		t.Parallel()

		recipes := newRecipesAPI(recipetest.NewMemoryStore())

		rec := recipes.do(t, http.MethodPost, "/api/recipes", alice, `{"name":"Idea"}`)
		require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

		body := `{"name":"Idea","sourceUrl":"","ingredients":[],"steps":[],"notes":"","archived":false,` +
			`"nutrition":{"kcal":{"amount":0,"missing":[]},"nutrients":{}}}`
		assert.JSONEq(t, body, withoutFields(t, rec.Body.String(), "id"))
	})

	t.Run("update", func(t *testing.T) {
		t.Parallel()

		recipes := newRecipesAPI(recipetest.NewMemoryStore())
		created := recipes.createSandwich(t)

		// The ID in the body is ignored in favour of the one in the path.
		changed := `{"id":"other","name":"Toast","steps":["Toast the bread."]}`

		rec := recipes.do(t, http.MethodPut, "/api/recipes/"+created.ID, alice, changed)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

		updated := decodeRecipe(t, rec)
		assert.Equal(t, created.ID, updated.ID)
		assert.Equal(t, "Toast", updated.Name)
		assert.Empty(t, updated.Ingredients)
		assert.Empty(t, updated.SourceURL)

		rec = recipes.do(t, http.MethodGet, "/api/recipes/"+created.ID, alice, "")
		assert.Equal(t, updated, decodeRecipe(t, rec))
	})

	t.Run("archive, unarchive and delete", func(t *testing.T) {
		t.Parallel()

		recipes := newRecipesAPI(recipetest.NewMemoryStore())
		created := recipes.createSandwich(t)
		path := "/api/recipes/" + created.ID
		breadID := created.Ingredients[0].FoodID

		rec := recipes.do(t, http.MethodDelete, path, alice, "")
		assert.Equal(t, http.StatusConflict, rec.Code)

		archived := strings.Replace(sandwich(breadID), `"archived":false`, `"archived":true`, 1)
		rec = recipes.do(t, http.MethodPut, path, alice, archived)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.True(t, decodeRecipe(t, rec).Archived)

		rec = recipes.do(t, http.MethodPut, path, alice, sandwich(breadID))
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.False(t, decodeRecipe(t, rec).Archived)

		rec = recipes.do(t, http.MethodPut, path, alice, archived)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

		rec = recipes.do(t, http.MethodDelete, path, alice, "")
		assert.Equal(t, http.StatusNoContent, rec.Code)

		rec = recipes.do(t, http.MethodGet, path, alice, "")
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("missing recipe", func(t *testing.T) {
		t.Parallel()

		recipes := newRecipesAPI(recipetest.NewMemoryStore())
		idea := `{"name":"Idea"}`

		assert.Equal(t, http.StatusNotFound, recipes.do(t, http.MethodGet, "/api/recipes/missing", alice, "").Code)
		assert.Equal(t, http.StatusNotFound, recipes.do(t, http.MethodPut, "/api/recipes/missing", alice, idea).Code)
		assert.Equal(t, http.StatusNotFound, recipes.do(t, http.MethodDelete, "/api/recipes/missing", alice, "").Code)
	})

	t.Run("users can't reach each other's recipes", func(t *testing.T) {
		t.Parallel()

		recipes := newRecipesAPI(recipetest.NewMemoryStore())
		alicesRecipe := recipes.createSandwich(t)
		path := "/api/recipes/" + alicesRecipe.ID

		assert.Equal(t, http.StatusNotFound, recipes.do(t, http.MethodGet, path, bob, "").Code)
		assert.Equal(t, http.StatusNotFound, recipes.do(t, http.MethodPut, path, bob, `{"name":"Mine"}`).Code)
		assert.Equal(t, http.StatusNotFound, recipes.do(t, http.MethodDelete, path, bob, "").Code)

		rec := recipes.do(t, http.MethodGet, "/api/recipes", bob, "")
		assert.JSONEq(t, "[]", rec.Body.String())

		rec = recipes.do(t, http.MethodGet, path, alice, "")
		assert.Equal(t, alicesRecipe, decodeRecipe(t, rec))
	})

	t.Run("not signed in", func(t *testing.T) {
		t.Parallel()

		rec := newRecipesAPI(recipetest.NewMemoryStore()).do(t, http.MethodGet, "/api/recipes", "", "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestRecipeIngredients(t *testing.T) {
	t.Parallel()

	t.Run("archived foods", func(t *testing.T) {
		t.Parallel()

		recipes := newRecipesAPI(recipetest.NewMemoryStore())
		bread := recipes.createBread(t)
		recipes.archiveFood(t, bread)

		rec := recipes.do(t, http.MethodPost, "/api/recipes", alice, sandwich(bread.ID))
		assert.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	})

	t.Run("unknown foods and servings", func(t *testing.T) {
		t.Parallel()

		recipes := newRecipesAPI(recipetest.NewMemoryStore())
		created := recipes.createSandwich(t)
		breadID := created.Ingredients[0].FoodID

		// A food that belongs to another user is as unknown as a missing one.
		rec := recipes.do(t, http.MethodPost, "/api/foods", bob, bread)
		require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

		bobsFood := decodeFood(t, rec)

		tests := []struct {
			name, body, wantMessage string
		}{
			{"missing food", sandwich("missing"), `there's no food with the ID "missing"`},
			{"another user's food", sandwich(bobsFood.ID), bobsFood.ID},
			{
				"unknown serving",
				strings.Replace(sandwich(breadID), `"serving":"slice"`, `"serving":"loaf"`, 1),
				`Rye bread has no serving called "loaf"`,
			},
		}

		for _, test := range tests {
			rec := recipes.do(t, http.MethodPost, "/api/recipes", alice, test.body)
			assert.Equal(t, http.StatusBadRequest, rec.Code, test.name)
			assert.Contains(t, rec.Body.String(), test.wantMessage, test.name)

			rec = recipes.do(t, http.MethodPut, "/api/recipes/"+created.ID, alice, test.body)
			assert.Equal(t, http.StatusBadRequest, rec.Code, test.name)
		}

		rec = recipes.do(t, http.MethodGet, "/api/recipes/"+created.ID, alice, "")
		assert.Equal(t, created, decodeRecipe(t, rec))
	})

	t.Run("foods can't be listed", func(t *testing.T) {
		t.Parallel()

		store := recipetest.NewMemoryStore()
		created, err := store.Create(t.Context(), alice, recipe.Recipe{Name: "Idea"})
		require.NoError(t, err)

		recipes := newAPI(server.Stores{Foods: failingFoodStore{}, Recipes: store})
		path := "/api/recipes/" + created.ID

		requests := []struct{ method, path, body string }{
			{http.MethodGet, "/api/recipes", ""},
			{http.MethodPost, "/api/recipes", sandwich("bread")},
			{http.MethodGet, path, ""},
			{http.MethodPut, path, sandwich("bread")},
		}

		for _, request := range requests {
			rec := recipes.do(t, request.method, request.path, alice, request.body)
			assert.Equal(t, http.StatusInternalServerError, rec.Code, request.method+" "+request.path)
		}
	})
}

func TestRecipeNutrition(t *testing.T) {
	t.Parallel()

	api := newRecipesAPI(recipetest.NewMemoryStore())
	proteinID := api.createProtein(t).ID

	// Bread has 8.5 g of protein per 100 g, while the protein in butter isn't
	// known.
	breadBody := strings.Replace(bread, `"nutrients":{}`, `"nutrients":{"`+proteinID+`":8.5}`, 1)
	rec := api.do(t, http.MethodPost, "/api/foods", alice, breadBody)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	breadID := decodeFood(t, rec).ID

	butterBody := `{"name":"Butter","unit":"g","kcalPer100":740,"servings":[]}`
	rec = api.do(t, http.MethodPost, "/api/foods", alice, butterBody)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	butterID := decodeFood(t, rec).ID

	body := `{"name":"Buttered toast","ingredients":[` +
		`{"foodId":"` + breadID + `","serving":"slice","amount":2},` +
		`{"foodId":"` + butterID + `","amount":10}]}`
	rec = api.do(t, http.MethodPost, "/api/recipes", alice, body)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	created := decodeRecipe(t, rec)

	want := nutritionJSON{
		// 210 kcal per 100 g * 80 g + 740 kcal per 100 g * 10 g
		Kcal: totalJSON{Amount: 242, Missing: []int{}},
		// 8.5 g per 100 g * 80 g, with butter's protein unknown.
		Nutrients: map[string]totalJSON{proteinID: {Amount: 6.8, Missing: []int{1}}},
	}

	assertNutrition(t, want, created.Nutrition)

	// The nutrition is worked out from the foods' current values.
	rec = api.do(t, http.MethodPut, "/api/foods/"+butterID, alice,
		`{"name":"Butter","unit":"g","kcalPer100":720,"servings":[],"nutrients":{"`+proteinID+`":1}}`)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	want = nutritionJSON{
		Kcal:      totalJSON{Amount: 240, Missing: []int{}},
		Nutrients: map[string]totalJSON{proteinID: {Amount: 6.9, Missing: []int{}}},
	}

	rec = api.do(t, http.MethodGet, "/api/recipes/"+created.ID, alice, "")
	require.Equal(t, http.StatusOK, rec.Code)
	assertNutrition(t, want, decodeRecipe(t, rec).Nutrition)

	rec = api.do(t, http.MethodGet, "/api/recipes", alice, "")
	require.Equal(t, http.StatusOK, rec.Code)

	var list []recipeJSON
	decode(t, rec, &list)
	require.Len(t, list, 1)
	assertNutrition(t, want, list[0].Nutrition)
}

func TestFoodsUsedByRecipes(t *testing.T) {
	t.Parallel()

	t.Run("changes that keep recipes' amounts", func(t *testing.T) {
		t.Parallel()

		api := newRecipesAPI(recipetest.NewMemoryStore())
		breadID := api.createSandwich(t).Ingredients[0].FoodID

		// Resizing a serving, adding another and archiving the food keep the
		// meaning of the sandwich's amounts.
		changed := strings.Replace(bread, `{"name":"slice","amount":40}`,
			`{"name":"slice","amount":45},{"name":"loaf","amount":750}`, 1)
		changed = strings.Replace(changed, `"archived":false`, `"archived":true`, 1)

		rec := api.do(t, http.MethodPut, "/api/foods/"+breadID, alice, changed)
		assert.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	})

	t.Run("changes that would change recipes' amounts", func(t *testing.T) {
		t.Parallel()

		api := newRecipesAPI(recipetest.NewMemoryStore())
		breadID := api.createSandwich(t).Ingredients[0].FoodID

		tests := []struct{ name, body, wantMessage string }{
			{
				"change the unit",
				strings.Replace(bread, `"unit":"g"`, `"unit":"ml"`, 1),
				`the unit can't be changed, because the food is used in "Open sandwich"`,
			},
			{
				"rename a serving",
				strings.Replace(bread, `"name":"slice"`, `"name":"thick slice"`, 1),
				`the serving "slice" can't be renamed or removed, because it's used in "Open sandwich"`,
			},
			{
				"remove a serving",
				strings.Replace(bread, `[{"name":"slice","amount":40}]`, `[]`, 1),
				`the serving "slice" can't be renamed or removed`,
			},
		}

		for _, test := range tests {
			rec := api.do(t, http.MethodPut, "/api/foods/"+breadID, alice, test.body)
			assert.Equal(t, http.StatusConflict, rec.Code, test.name)
			assert.Contains(t, rec.Body.String(), test.wantMessage, test.name)
		}

		rec := api.do(t, http.MethodGet, "/api/foods/"+breadID, alice, "")
		assert.JSONEq(t, bread, withoutID(t, rec.Body.String()))
	})

	t.Run("delete", func(t *testing.T) {
		t.Parallel()

		api := newRecipesAPI(recipetest.NewMemoryStore())
		created := api.createSandwich(t)
		breadID := created.Ingredients[0].FoodID
		api.archiveFood(t, foodJSON{ID: breadID})

		rec := api.do(t, http.MethodDelete, "/api/foods/"+breadID, alice, "")
		assert.Equal(t, http.StatusConflict, rec.Code)
		assert.Contains(t, rec.Body.String(), `the food can't be deleted, because it's used in "Open sandwich"`)

		// Archived recipes still use the food, so deleting the recipe is what
		// allows deleting the food.
		archived := strings.Replace(sandwich(breadID), `"archived":false`, `"archived":true`, 1)
		rec = api.do(t, http.MethodPut, "/api/recipes/"+created.ID, alice, archived)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

		rec = api.do(t, http.MethodDelete, "/api/foods/"+breadID, alice, "")
		assert.Equal(t, http.StatusConflict, rec.Code)

		rec = api.do(t, http.MethodDelete, "/api/recipes/"+created.ID, alice, "")
		require.Equal(t, http.StatusNoContent, rec.Code)

		rec = api.do(t, http.MethodDelete, "/api/foods/"+breadID, alice, "")
		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("recipes can't be listed", func(t *testing.T) {
		t.Parallel()

		foods := foodtest.NewMemoryStore()
		created, err := foods.Create(t.Context(), alice, foodtest.Bread())
		require.NoError(t, err)

		api := newAPI(server.Stores{
			Foods:     foods,
			Nutrients: nutrienttest.NewMemoryStore(),
			Recipes:   failingRecipeStore{},
		})
		path := "/api/foods/" + created.ID

		assert.Equal(t, http.StatusInternalServerError, api.do(t, http.MethodPut, path, alice, bread).Code)
		assert.Equal(t, http.StatusInternalServerError, api.do(t, http.MethodDelete, path, alice, "").Code)
	})
}

// assertNutrition checks that got matches want, allowing for rounding errors in
// the amounts.
func assertNutrition(t *testing.T, want, got nutritionJSON) {
	t.Helper()

	assert.InDelta(t, want.Kcal.Amount, got.Kcal.Amount, 1e-9)
	assert.Equal(t, want.Kcal.Missing, got.Kcal.Missing)
	assert.Len(t, got.Nutrients, len(want.Nutrients))

	for id, total := range want.Nutrients {
		assert.InDelta(t, total.Amount, got.Nutrients[id].Amount, 1e-9, id)
		assert.Equal(t, total.Missing, got.Nutrients[id].Missing, id)
	}
}

func TestRecipesInvalidInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
	}{
		{"not JSON", "soup"},
		{"wrong type", `{"name":"Soup","steps":"Boil."}`},
		{"unknown field", `{"name":"Soup","servings":4}`},
		{"nutrition", `{"name":"Soup","nutrition":{"kcal":{"amount":100,"missing":[]},"nutrients":{}}}`},
		{"invalid recipe", `{"name":""}`},
		{"invalid ingredient", `{"name":"Soup","ingredients":[{"foodId":"bread","amount":0}]}`},
		{"too large", `{"name":"` + strings.Repeat("a", 1<<20) + `"}`},
	}

	recipes := newRecipesAPI(recipetest.NewMemoryStore())
	created := recipes.createSandwich(t)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			rec := recipes.do(t, http.MethodPost, "/api/recipes", alice, test.body)
			assert.Equal(t, http.StatusBadRequest, rec.Code)

			rec = recipes.do(t, http.MethodPut, "/api/recipes/"+created.ID, alice, test.body)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})
	}
}

// failingRecipeStore is a recipe.Store whose every method fails.
type failingRecipeStore struct {
	recipe.Store
}

func (failingRecipeStore) Create(context.Context, string, recipe.Recipe) (recipe.Recipe, error) {
	return recipe.Recipe{}, errStore
}

func (failingRecipeStore) Get(context.Context, string, string) (recipe.Recipe, error) {
	return recipe.Recipe{}, errStore
}

func (failingRecipeStore) List(context.Context, string) ([]recipe.Recipe, error) {
	return nil, errStore
}

func (failingRecipeStore) Update(context.Context, string, recipe.Recipe) (recipe.Recipe, error) {
	return recipe.Recipe{}, errStore
}

func (failingRecipeStore) Delete(context.Context, string, string) error {
	return errStore
}

func TestRecipesStoreError(t *testing.T) {
	t.Parallel()

	recipes := newRecipesAPI(failingRecipeStore{})

	requests := []struct{ method, path, body string }{
		{http.MethodGet, "/api/recipes", ""},
		{http.MethodPost, "/api/recipes", `{"name":"Idea"}`},
		{http.MethodGet, "/api/recipes/some-id", ""},
		{http.MethodPut, "/api/recipes/some-id", `{"name":"Idea"}`},
		{http.MethodDelete, "/api/recipes/some-id", ""},
	}

	for _, request := range requests {
		rec := recipes.do(t, request.method, request.path, alice, request.body)
		assert.Equal(t, http.StatusInternalServerError, rec.Code, request.method+" "+request.path)
		assert.NotContains(t, rec.Body.String(), errStore.Error())
	}
}
