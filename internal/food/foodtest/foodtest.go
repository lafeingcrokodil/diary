// Package foodtest checks that implementations of food.Store behave the same
// way, so that each implementation's tests can run the same checks.
package foodtest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lafeingcrokodil/diary/internal/food"
)

const (
	alice = "alice"
	bob   = "bob"
)

// Bread returns a valid food, which tests can change as they need.
func Bread() food.Food {
	return food.Food{
		Name:       "Rye bread",
		Brand:      "Bakery",
		Unit:       food.Grams,
		KcalPer100: 210,
		Servings:   []food.Serving{{Name: "slice", Amount: 40}, {Name: "loaf", Amount: 750}},
		Nutrients:  map[string]float64{"protein": 8.5, "fibre": 6},
	}
}

func archivedBread() food.Food {
	item := Bread()
	item.Archived = true

	return item
}

// CheckStore checks that a Store implementation behaves as the interface
// describes. newStore must return an empty store each time it's called. It
// gets the test that the store is for, so that it can fail that test if it
// can't create a store.
func CheckStore(t *testing.T, newStore func(t *testing.T) food.Store) {
	t.Helper()

	t.Run("create and get", func(t *testing.T) {
		t.Parallel()

		store := newStore(t)

		input := Bread()
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

		first, err := store.Create(t.Context(), alice, Bread())
		require.NoError(t, err)

		second, err := store.Create(t.Context(), alice, Bread())
		require.NoError(t, err)

		assert.NotEqual(t, first.ID, second.ID)
	})

	t.Run("get missing food", func(t *testing.T) {
		t.Parallel()

		_, err := newStore(t).Get(t.Context(), alice, "missing")
		require.ErrorIs(t, err, food.ErrNotFound)
	})

	t.Run("unusual IDs don't match any food", func(t *testing.T) {
		t.Parallel()

		store := newStore(t)

		created, err := store.Create(t.Context(), alice, Bread())
		require.NoError(t, err)

		for _, id := range []string{"", "/", created.ID + "/", "foods/" + created.ID} {
			_, err := store.Get(t.Context(), alice, id)
			require.ErrorIs(t, err, food.ErrNotFound, "Get(%q)", id)

			item := Bread()
			item.ID = id
			_, err = store.Update(t.Context(), alice, item)
			require.ErrorIs(t, err, food.ErrNotFound, "Update(%q)", id)

			require.ErrorIs(t, store.Delete(t.Context(), alice, id), food.ErrNotFound, "Delete(%q)", id)
		}
	})

	t.Run("list sorts by name", func(t *testing.T) {
		t.Parallel()

		store := newStore(t)

		for _, name := range []string{"rye bread", "Apple", "Butter"} {
			f := Bread()
			f.Name = name
			_, err := store.Create(t.Context(), alice, f)
			require.NoError(t, err)
		}

		foods, err := store.List(t.Context(), alice)
		require.NoError(t, err)

		names := make([]string, 0, len(foods))
		for _, f := range foods {
			names = append(names, f.Name)
		}

		assert.Equal(t, []string{"Apple", "Butter", "rye bread"}, names)
	})

	t.Run("list without foods", func(t *testing.T) {
		t.Parallel()

		foods, err := newStore(t).List(t.Context(), alice)
		require.NoError(t, err)
		assert.Empty(t, foods)
	})

	t.Run("update", func(t *testing.T) {
		t.Parallel()

		store := newStore(t)

		created, err := store.Create(t.Context(), alice, Bread())
		require.NoError(t, err)

		changed := created
		changed.Name = "Wholemeal rye bread"
		changed.Servings = []food.Serving{{Name: "thin slice", Amount: 25}}
		changed.Nutrients = map[string]float64{"protein": 9}
		changed.Archived = true
		changed.Owner = bob

		updated, err := store.Update(t.Context(), alice, changed)
		require.NoError(t, err)
		assert.Equal(t, alice, updated.Owner)
		assert.Equal(t, "Wholemeal rye bread", updated.Name)

		got, err := store.Get(t.Context(), alice, created.ID)
		require.NoError(t, err)
		assert.Equal(t, updated, got)
	})

	t.Run("update missing food", func(t *testing.T) {
		t.Parallel()

		f := Bread()
		f.ID = "missing"

		_, err := newStore(t).Update(t.Context(), alice, f)
		require.ErrorIs(t, err, food.ErrNotFound)
	})

	t.Run("delete", func(t *testing.T) {
		t.Parallel()

		store := newStore(t)

		created, err := store.Create(t.Context(), alice, archivedBread())
		require.NoError(t, err)

		require.NoError(t, store.Delete(t.Context(), alice, created.ID))

		_, err = store.Get(t.Context(), alice, created.ID)
		require.ErrorIs(t, err, food.ErrNotFound)

		require.ErrorIs(t, store.Delete(t.Context(), alice, created.ID), food.ErrNotFound)
	})

	t.Run("delete a food that isn't archived", func(t *testing.T) {
		t.Parallel()

		store := newStore(t)

		created, err := store.Create(t.Context(), alice, Bread())
		require.NoError(t, err)

		require.ErrorIs(t, store.Delete(t.Context(), alice, created.ID), food.ErrNotArchived)

		got, err := store.Get(t.Context(), alice, created.ID)
		require.NoError(t, err)
		assert.Equal(t, created, got)
	})

	t.Run("users can't reach each other's foods", func(t *testing.T) {
		t.Parallel()

		store := newStore(t)

		// Archived, so that only the owner check stops Bob from deleting it.
		alicesFood, err := store.Create(t.Context(), alice, archivedBread())
		require.NoError(t, err)

		_, err = store.Get(t.Context(), bob, alicesFood.ID)
		require.ErrorIs(t, err, food.ErrNotFound)

		bobsFoods, err := store.List(t.Context(), bob)
		require.NoError(t, err)
		assert.Empty(t, bobsFoods)

		changed := alicesFood
		changed.Name = "Changed by Bob"
		_, err = store.Update(t.Context(), bob, changed)
		require.ErrorIs(t, err, food.ErrNotFound)

		require.ErrorIs(t, store.Delete(t.Context(), bob, alicesFood.ID), food.ErrNotFound)

		got, err := store.Get(t.Context(), alice, alicesFood.ID)
		require.NoError(t, err)
		assert.Equal(t, alicesFood, got)
	})

	t.Run("foods without nutrients", func(t *testing.T) {
		t.Parallel()

		store := newStore(t)

		input := Bread()
		input.Nutrients = nil

		created, err := store.Create(t.Context(), alice, input)
		require.NoError(t, err)

		got, err := store.Get(t.Context(), alice, created.ID)
		require.NoError(t, err)
		assert.Empty(t, got.Nutrients)
	})

	t.Run("callers don't share nutrients with the store", func(t *testing.T) {
		t.Parallel()

		store := newStore(t)

		input := Bread()
		created, err := store.Create(t.Context(), alice, input)
		require.NoError(t, err)

		input.Nutrients["protein"] = 1
		created.Nutrients["protein"] = 2

		got, err := store.Get(t.Context(), alice, created.ID)
		require.NoError(t, err)
		assert.InDelta(t, 8.5, got.Nutrients["protein"], 0)

		got.Nutrients["protein"] = 3

		again, err := store.Get(t.Context(), alice, created.ID)
		require.NoError(t, err)
		assert.InDelta(t, 8.5, again.Nutrients["protein"], 0)
	})

	t.Run("callers don't share servings with the store", func(t *testing.T) {
		t.Parallel()

		store := newStore(t)

		input := Bread()
		created, err := store.Create(t.Context(), alice, input)
		require.NoError(t, err)

		input.Servings[0].Name = "changed input"
		created.Servings[0].Name = "changed result"

		got, err := store.Get(t.Context(), alice, created.ID)
		require.NoError(t, err)
		assert.Equal(t, "slice", got.Servings[0].Name)
	})
}
