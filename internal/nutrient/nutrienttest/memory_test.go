package nutrienttest_test

import (
	"testing"

	"github.com/lafeingcrokodil/diary/internal/food/foodtest"
	"github.com/lafeingcrokodil/diary/internal/nutrient"
	"github.com/lafeingcrokodil/diary/internal/nutrient/nutrienttest"
)

func TestMemoryStore(t *testing.T) {
	t.Parallel()

	nutrienttest.CheckStore(t, func(*testing.T) nutrient.Store { return nutrienttest.NewMemoryStore() })
}

func TestMemoryStoreDeleteRemovesAmounts(t *testing.T) {
	t.Parallel()

	nutrienttest.CheckDeleteRemovesAmounts(t, func(*testing.T) nutrienttest.Stores {
		foods := foodtest.NewMemoryStore()

		return nutrienttest.Stores{Foods: foods, Nutrients: nutrienttest.NewMemoryStoreWithFoods(foods)}
	})
}
