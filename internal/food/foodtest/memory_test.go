package foodtest_test

import (
	"testing"

	"github.com/lafeingcrokodil/diary/internal/food"
	"github.com/lafeingcrokodil/diary/internal/food/foodtest"
)

func TestMemoryStore(t *testing.T) {
	t.Parallel()

	foodtest.CheckStore(t, func() food.Store { return foodtest.NewMemoryStore() })
}
