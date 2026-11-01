package recipetest_test

import (
	"testing"

	"github.com/lafeingcrokodil/diary/internal/recipe"
	"github.com/lafeingcrokodil/diary/internal/recipe/recipetest"
)

func TestMemoryStore(t *testing.T) {
	t.Parallel()

	recipetest.CheckStore(t, func(*testing.T) recipe.Store { return recipetest.NewMemoryStore() })
}
