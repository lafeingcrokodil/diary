package nutrienttest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lafeingcrokodil/diary/internal/food"
	"github.com/lafeingcrokodil/diary/internal/food/foodtest"
	"github.com/lafeingcrokodil/diary/internal/nutrient"
)

// Stores is a food store and a nutrient store that share the same data.
type Stores struct {
	Foods     food.Store
	Nutrients nutrient.Store
}

// CheckDeleteRemovesAmounts checks that deleting a nutrient removes its
// amounts from its owner's foods, and only then. newStores must return empty
// stores each time it's called.
func CheckDeleteRemovesAmounts(t *testing.T, newStores func(t *testing.T) Stores) {
	t.Helper()

	// setUp creates two of alice's nutrients, the first one archived, and a
	// food of alice's and one of bob's with amounts of both.
	setUp := func(t *testing.T) (Stores, []nutrient.Nutrient, []food.Food) {
		t.Helper()

		stores := newStores(t)

		archived, err := stores.Nutrients.Create(t.Context(), alice, archivedProtein())
		require.NoError(t, err)

		active, err := stores.Nutrients.Create(t.Context(), alice, Protein())
		require.NoError(t, err)

		owners := []string{alice, bob}
		foods := make([]food.Food, 0, len(owners))

		for _, owner := range owners {
			item := foodtest.Bread()
			item.Nutrients = map[string]float64{archived.ID: 8.5, active.ID: 6}

			created, err := stores.Foods.Create(t.Context(), owner, item)
			require.NoError(t, err)

			foods = append(foods, created)
		}

		return stores, []nutrient.Nutrient{archived, active}, foods
	}

	// amounts returns the nutrient amounts of one of owner's foods.
	amounts := func(t *testing.T, stores Stores, owner string, item food.Food) map[string]float64 {
		t.Helper()

		got, err := stores.Foods.Get(t.Context(), owner, item.ID)
		require.NoError(t, err)

		return got.Nutrients
	}

	t.Run("deleting a nutrient removes its amounts from its owner's foods", func(t *testing.T) {
		t.Parallel()

		stores, nutrients, foods := setUp(t)
		archived, active := nutrients[0], nutrients[1]

		require.NoError(t, stores.Nutrients.Delete(t.Context(), alice, archived.ID))

		assert.Equal(t, map[string]float64{active.ID: 6}, amounts(t, stores, alice, foods[0]))

		// Bob's food can't have amounts of Alice's nutrients through the API,
		// but if it did, they'd be left alone.
		assert.Equal(t, map[string]float64{archived.ID: 8.5, active.ID: 6}, amounts(t, stores, bob, foods[1]))
	})

	t.Run("a nutrient that can't be deleted keeps its amounts", func(t *testing.T) {
		t.Parallel()

		stores, nutrients, foods := setUp(t)
		archived, active := nutrients[0], nutrients[1]

		require.ErrorIs(t, stores.Nutrients.Delete(t.Context(), alice, active.ID), nutrient.ErrNotArchived)
		require.ErrorIs(t, stores.Nutrients.Delete(t.Context(), bob, archived.ID), nutrient.ErrNotFound)

		assert.Equal(t, map[string]float64{archived.ID: 8.5, active.ID: 6}, amounts(t, stores, alice, foods[0]))
	})
}
