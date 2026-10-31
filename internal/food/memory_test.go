package food_test

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

func TestMemoryStore(t *testing.T) {
	t.Parallel()

	testStore(t, func() food.Store { return food.NewMemoryStore() })
}

// testStore checks that a Store implementation behaves as the interface
// describes. newStore must return an empty store each time it's called.
func testStore(t *testing.T, newStore func() food.Store) {
	t.Helper()

	t.Run("create and get", func(t *testing.T) {
		t.Parallel()

		store := newStore()

		input := bread()
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

		first, err := store.Create(t.Context(), alice, bread())
		require.NoError(t, err)

		second, err := store.Create(t.Context(), alice, bread())
		require.NoError(t, err)

		assert.NotEqual(t, first.ID, second.ID)
	})

	t.Run("get missing food", func(t *testing.T) {
		t.Parallel()

		_, err := newStore().Get(t.Context(), alice, "missing")
		require.ErrorIs(t, err, food.ErrNotFound)
	})

	t.Run("list sorts by name", func(t *testing.T) {
		t.Parallel()

		store := newStore()

		for _, name := range []string{"rye bread", "Apple", "Butter"} {
			f := bread()
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

		foods, err := newStore().List(t.Context(), alice)
		require.NoError(t, err)
		assert.Empty(t, foods)
	})

	t.Run("update", func(t *testing.T) {
		t.Parallel()

		store := newStore()

		created, err := store.Create(t.Context(), alice, bread())
		require.NoError(t, err)

		changed := created
		changed.Name = "Wholemeal rye bread"
		changed.Servings = []food.Serving{{Name: "thin slice", Amount: 25}}
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

		f := bread()
		f.ID = "missing"

		_, err := newStore().Update(t.Context(), alice, f)
		require.ErrorIs(t, err, food.ErrNotFound)
	})

	t.Run("delete", func(t *testing.T) {
		t.Parallel()

		store := newStore()

		created, err := store.Create(t.Context(), alice, bread())
		require.NoError(t, err)

		require.NoError(t, store.Delete(t.Context(), alice, created.ID))

		_, err = store.Get(t.Context(), alice, created.ID)
		require.ErrorIs(t, err, food.ErrNotFound)

		require.ErrorIs(t, store.Delete(t.Context(), alice, created.ID), food.ErrNotFound)
	})

	t.Run("users can't reach each other's foods", func(t *testing.T) {
		t.Parallel()

		store := newStore()

		alicesFood, err := store.Create(t.Context(), alice, bread())
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

	t.Run("callers don't share servings with the store", func(t *testing.T) {
		t.Parallel()

		store := newStore()

		input := bread()
		created, err := store.Create(t.Context(), alice, input)
		require.NoError(t, err)

		input.Servings[0].Name = "changed input"
		created.Servings[0].Name = "changed result"

		got, err := store.Get(t.Context(), alice, created.ID)
		require.NoError(t, err)
		assert.Equal(t, "slice", got.Servings[0].Name)
	})
}
