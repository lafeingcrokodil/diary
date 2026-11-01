// Package nutrienttest checks that implementations of nutrient.Store behave
// the same way, so that each implementation's tests can run the same checks.
package nutrienttest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lafeingcrokodil/diary/internal/nutrient"
)

const (
	alice = "alice"
	bob   = "bob"
)

// Protein returns a valid nutrient, which tests can change as they need.
func Protein() nutrient.Nutrient {
	return nutrient.Nutrient{
		Name:        "Protein",
		Unit:        nutrient.Grams,
		KcalPerGram: 4,
	}
}

// CheckStore checks that a Store implementation behaves as the interface
// describes. newStore must return an empty store each time it's called.
func CheckStore(t *testing.T, newStore func() nutrient.Store) {
	t.Helper()

	t.Run("create and get", func(t *testing.T) {
		t.Parallel()

		store := newStore()

		input := Protein()
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

		store := newStore()

		first, err := store.Create(t.Context(), alice, Protein())
		require.NoError(t, err)

		second, err := store.Create(t.Context(), alice, Protein())
		require.NoError(t, err)

		assert.NotEqual(t, first.ID, second.ID)
	})

	t.Run("get missing nutrient", func(t *testing.T) {
		t.Parallel()

		_, err := newStore().Get(t.Context(), alice, "missing")
		require.ErrorIs(t, err, nutrient.ErrNotFound)
	})

	t.Run("unusual IDs don't match any nutrient", func(t *testing.T) {
		t.Parallel()

		store := newStore()

		created, err := store.Create(t.Context(), alice, Protein())
		require.NoError(t, err)

		for _, id := range []string{"", "/", created.ID + "/", "nutrients/" + created.ID} {
			_, err := store.Get(t.Context(), alice, id)
			require.ErrorIs(t, err, nutrient.ErrNotFound, "Get(%q)", id)

			item := Protein()
			item.ID = id
			_, err = store.Update(t.Context(), alice, item)
			require.ErrorIs(t, err, nutrient.ErrNotFound, "Update(%q)", id)

			require.ErrorIs(t, store.Delete(t.Context(), alice, id), nutrient.ErrNotFound, "Delete(%q)", id)
		}
	})

	t.Run("list sorts by name and includes archived nutrients", func(t *testing.T) {
		t.Parallel()

		store := newStore()

		for _, name := range []string{"protein", "Carbohydrates", "Fibre"} {
			n := Protein()
			n.Name = name
			n.Archived = name == "Fibre"
			_, err := store.Create(t.Context(), alice, n)
			require.NoError(t, err)
		}

		nutrients, err := store.List(t.Context(), alice)
		require.NoError(t, err)

		names := make([]string, 0, len(nutrients))
		for _, n := range nutrients {
			names = append(names, n.Name)
		}

		assert.Equal(t, []string{"Carbohydrates", "Fibre", "protein"}, names)
	})

	t.Run("list without nutrients", func(t *testing.T) {
		t.Parallel()

		nutrients, err := newStore().List(t.Context(), alice)
		require.NoError(t, err)
		assert.Empty(t, nutrients)
	})

	t.Run("update", func(t *testing.T) {
		t.Parallel()

		store := newStore()

		created, err := store.Create(t.Context(), alice, Protein())
		require.NoError(t, err)

		changed := created
		changed.Name = "Plant protein"
		changed.Unit = nutrient.Milligrams
		changed.KcalPerGram = 0
		changed.Archived = true
		changed.Owner = bob

		updated, err := store.Update(t.Context(), alice, changed)
		require.NoError(t, err)

		changed.Owner = alice
		assert.Equal(t, changed, updated)

		got, err := store.Get(t.Context(), alice, created.ID)
		require.NoError(t, err)
		assert.Equal(t, updated, got)
	})

	t.Run("update missing nutrient", func(t *testing.T) {
		t.Parallel()

		n := Protein()
		n.ID = "missing"

		_, err := newStore().Update(t.Context(), alice, n)
		require.ErrorIs(t, err, nutrient.ErrNotFound)
	})

	t.Run("delete", func(t *testing.T) {
		t.Parallel()

		store := newStore()

		created, err := store.Create(t.Context(), alice, Protein())
		require.NoError(t, err)

		require.NoError(t, store.Delete(t.Context(), alice, created.ID))

		_, err = store.Get(t.Context(), alice, created.ID)
		require.ErrorIs(t, err, nutrient.ErrNotFound)

		require.ErrorIs(t, store.Delete(t.Context(), alice, created.ID), nutrient.ErrNotFound)
	})

	t.Run("users can't reach each other's nutrients", func(t *testing.T) {
		t.Parallel()

		store := newStore()

		alicesNutrient, err := store.Create(t.Context(), alice, Protein())
		require.NoError(t, err)

		_, err = store.Get(t.Context(), bob, alicesNutrient.ID)
		require.ErrorIs(t, err, nutrient.ErrNotFound)

		bobsNutrients, err := store.List(t.Context(), bob)
		require.NoError(t, err)
		assert.Empty(t, bobsNutrients)

		changed := alicesNutrient
		changed.Name = "Changed by Bob"
		_, err = store.Update(t.Context(), bob, changed)
		require.ErrorIs(t, err, nutrient.ErrNotFound)

		require.ErrorIs(t, store.Delete(t.Context(), bob, alicesNutrient.ID), nutrient.ErrNotFound)

		got, err := store.Get(t.Context(), alice, alicesNutrient.ID)
		require.NoError(t, err)
		assert.Equal(t, alicesNutrient, got)
	})
}
