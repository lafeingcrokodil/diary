// Package recipetest checks that implementations of recipe.Store behave the
// same way, so that each implementation's tests can run the same checks.
package recipetest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lafeingcrokodil/diary/internal/recipe"
)

const (
	alice = "alice"
	bob   = "bob"
)

// Soup returns a valid recipe, which tests can change as they need. Its
// ingredients refer to foods by made-up IDs, since the store doesn't check
// them.
func Soup() recipe.Recipe {
	return recipe.Recipe{
		Name:      "Lentil soup",
		SourceURL: "https://example.com/lentil-soup",
		Ingredients: []recipe.Ingredient{
			{FoodID: "lentils", Amount: 200, Note: "rinsed"},
			{FoodID: "onion", Serving: "medium onion", Amount: 1, Note: "finely chopped"},
		},
		Steps: []string{"Fry the onion.", "Add the lentils and simmer for 20 minutes."},
		Notes: "Freezes well.",
	}
}

func archivedSoup() recipe.Recipe {
	item := Soup()
	item.Archived = true

	return item
}

// CheckStore checks that a Store implementation behaves as the interface
// describes. newStore must return an empty store each time it's called. It
// gets the test that the store is for, so that it can fail that test if it
// can't create a store.
func CheckStore(t *testing.T, newStore func(t *testing.T) recipe.Store) {
	t.Helper()

	t.Run("create and get", func(t *testing.T) {
		t.Parallel()

		store := newStore(t)

		input := Soup()
		input.ID = "ignored"
		input.Owner = bob

		created, err := store.Create(t.Context(), alice, input)
		require.NoError(t, err)
		assert.NotEmpty(t, created.ID)
		assert.NotEqual(t, "ignored", created.ID)
		assert.Equal(t, alice, created.Owner)

		got, err := store.Get(t.Context(), alice, created.ID)
		require.NoError(t, err)
		assert.Equal(t, created, got)
	})

	t.Run("IDs are unique", func(t *testing.T) {
		t.Parallel()

		store := newStore(t)

		first, err := store.Create(t.Context(), alice, Soup())
		require.NoError(t, err)

		second, err := store.Create(t.Context(), alice, Soup())
		require.NoError(t, err)

		assert.NotEqual(t, first.ID, second.ID)
	})

	t.Run("get missing recipe", func(t *testing.T) {
		t.Parallel()

		_, err := newStore(t).Get(t.Context(), alice, "missing")
		require.ErrorIs(t, err, recipe.ErrNotFound)
	})

	t.Run("unusual IDs don't match any recipe", func(t *testing.T) {
		t.Parallel()

		store := newStore(t)

		created, err := store.Create(t.Context(), alice, Soup())
		require.NoError(t, err)

		for _, id := range []string{"", "/", created.ID + "/", "recipes/" + created.ID} {
			_, err := store.Get(t.Context(), alice, id)
			require.ErrorIs(t, err, recipe.ErrNotFound, "Get(%q)", id)

			item := Soup()
			item.ID = id
			_, err = store.Update(t.Context(), alice, item)
			require.ErrorIs(t, err, recipe.ErrNotFound, "Update(%q)", id)

			require.ErrorIs(t, store.Delete(t.Context(), alice, id), recipe.ErrNotFound, "Delete(%q)", id)
		}
	})

	t.Run("list sorts by name", func(t *testing.T) {
		t.Parallel()

		store := newStore(t)

		for _, name := range []string{"lentil soup", "Apple pie", "Bread"} {
			r := Soup()
			r.Name = name
			_, err := store.Create(t.Context(), alice, r)
			require.NoError(t, err)
		}

		recipes, err := store.List(t.Context(), alice)
		require.NoError(t, err)

		names := make([]string, 0, len(recipes))
		for _, r := range recipes {
			names = append(names, r.Name)
		}

		assert.Equal(t, []string{"Apple pie", "Bread", "lentil soup"}, names)
	})

	t.Run("list without recipes", func(t *testing.T) {
		t.Parallel()

		recipes, err := newStore(t).List(t.Context(), alice)
		require.NoError(t, err)
		assert.Empty(t, recipes)
	})

	t.Run("update", func(t *testing.T) {
		t.Parallel()

		store := newStore(t)

		created, err := store.Create(t.Context(), alice, Soup())
		require.NoError(t, err)

		changed := created
		changed.Name = "Spicy lentil soup"
		changed.SourceURL = ""
		changed.Ingredients = []recipe.Ingredient{{FoodID: "chilli", Serving: "chilli", Amount: 2}}
		changed.Steps = []string{"Chop the chillies."}
		changed.Notes = ""
		changed.Archived = true
		changed.Owner = bob

		updated, err := store.Update(t.Context(), alice, changed)
		require.NoError(t, err)
		assert.Equal(t, alice, updated.Owner)
		assert.Equal(t, "Spicy lentil soup", updated.Name)

		got, err := store.Get(t.Context(), alice, created.ID)
		require.NoError(t, err)
		assert.Equal(t, updated, got)
	})

	t.Run("update missing recipe", func(t *testing.T) {
		t.Parallel()

		r := Soup()
		r.ID = "missing"

		_, err := newStore(t).Update(t.Context(), alice, r)
		require.ErrorIs(t, err, recipe.ErrNotFound)
	})

	t.Run("delete", func(t *testing.T) {
		t.Parallel()

		store := newStore(t)

		created, err := store.Create(t.Context(), alice, archivedSoup())
		require.NoError(t, err)

		require.NoError(t, store.Delete(t.Context(), alice, created.ID))

		_, err = store.Get(t.Context(), alice, created.ID)
		require.ErrorIs(t, err, recipe.ErrNotFound)

		require.ErrorIs(t, store.Delete(t.Context(), alice, created.ID), recipe.ErrNotFound)
	})

	t.Run("delete a recipe that isn't archived", func(t *testing.T) {
		t.Parallel()

		store := newStore(t)

		created, err := store.Create(t.Context(), alice, Soup())
		require.NoError(t, err)

		require.ErrorIs(t, store.Delete(t.Context(), alice, created.ID), recipe.ErrNotArchived)

		got, err := store.Get(t.Context(), alice, created.ID)
		require.NoError(t, err)
		assert.Equal(t, created, got)
	})

	t.Run("users can't reach each other's recipes", func(t *testing.T) {
		t.Parallel()

		store := newStore(t)

		// Archived, so that only the owner check stops Bob from deleting it.
		alicesRecipe, err := store.Create(t.Context(), alice, archivedSoup())
		require.NoError(t, err)

		_, err = store.Get(t.Context(), bob, alicesRecipe.ID)
		require.ErrorIs(t, err, recipe.ErrNotFound)

		bobsRecipes, err := store.List(t.Context(), bob)
		require.NoError(t, err)
		assert.Empty(t, bobsRecipes)

		changed := alicesRecipe
		changed.Name = "Changed by Bob"
		_, err = store.Update(t.Context(), bob, changed)
		require.ErrorIs(t, err, recipe.ErrNotFound)

		require.ErrorIs(t, store.Delete(t.Context(), bob, alicesRecipe.ID), recipe.ErrNotFound)

		got, err := store.Get(t.Context(), alice, alicesRecipe.ID)
		require.NoError(t, err)
		assert.Equal(t, alicesRecipe, got)
	})

	t.Run("recipes without ingredients or steps", func(t *testing.T) {
		t.Parallel()

		store := newStore(t)

		input := Soup()
		input.Ingredients = nil
		input.Steps = nil

		created, err := store.Create(t.Context(), alice, input)
		require.NoError(t, err)

		got, err := store.Get(t.Context(), alice, created.ID)
		require.NoError(t, err)
		assert.Empty(t, got.Ingredients)
		assert.Empty(t, got.Steps)
	})

	t.Run("callers don't share ingredients or steps with the store", func(t *testing.T) {
		t.Parallel()

		store := newStore(t)

		input := Soup()
		created, err := store.Create(t.Context(), alice, input)
		require.NoError(t, err)

		input.Ingredients[0].Note = "changed input"
		input.Steps[0] = "changed input"
		created.Ingredients[0].Note = "changed result"
		created.Steps[0] = "changed result"

		got, err := store.Get(t.Context(), alice, created.ID)
		require.NoError(t, err)
		assert.Equal(t, Soup().Ingredients, got.Ingredients)
		assert.Equal(t, Soup().Steps, got.Steps)

		got.Ingredients[0].Note = "changed copy"
		got.Steps[0] = "changed copy"

		again, err := store.Get(t.Context(), alice, created.ID)
		require.NoError(t, err)
		assert.Equal(t, Soup().Ingredients, again.Ingredients)
		assert.Equal(t, Soup().Steps, again.Steps)
	})
}
