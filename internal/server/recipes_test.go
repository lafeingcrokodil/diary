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
	Steps    []string `json:"steps"`
	Notes    string   `json:"notes"`
	Archived bool     `json:"archived"`
}

// newRecipesAPI returns an api whose recipes are kept in store, and whose foods
// are kept in memory.
func newRecipesAPI(store recipe.Store) api {
	return newAPI(server.Stores{Foods: foodtest.NewMemoryStore(), Recipes: store})
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
		assert.JSONEq(t, sandwich(breadID), withoutID(t, rec.Body.String()))

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

		body := `{"name":"Idea","sourceUrl":"","ingredients":[],"steps":[],"notes":"","archived":false}`
		assert.JSONEq(t, body, withoutID(t, rec.Body.String()))
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

		recipes := newAPI(server.Stores{Foods: failingFoodStore{}, Recipes: recipetest.NewMemoryStore()})

		rec := recipes.do(t, http.MethodPost, "/api/recipes", alice, sandwich("bread"))
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
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
