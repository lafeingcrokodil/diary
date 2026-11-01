package food_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lafeingcrokodil/diary/internal/food"
)

func TestSortByName(t *testing.T) {
	t.Parallel()

	foods := []food.Food{
		{ID: "b", Name: "rye bread"},
		{ID: "c", Name: "Butter"},
		{ID: "a", Name: "rye bread"},
		{ID: "d", Name: "apple"},
	}

	food.SortByName(foods)

	ids := make([]string, 0, len(foods))
	for _, item := range foods {
		ids = append(ids, item.ID)
	}

	// Names are compared ignoring case, and equal names are ordered by ID.
	assert.Equal(t, []string{"d", "c", "a", "b"}, ids)
}
