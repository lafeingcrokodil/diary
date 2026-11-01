package nutrienttest_test

import (
	"testing"

	"github.com/lafeingcrokodil/diary/internal/nutrient"
	"github.com/lafeingcrokodil/diary/internal/nutrient/nutrienttest"
)

func TestMemoryStore(t *testing.T) {
	t.Parallel()

	nutrienttest.CheckStore(t, func(*testing.T) nutrient.Store { return nutrienttest.NewMemoryStore() })
}
