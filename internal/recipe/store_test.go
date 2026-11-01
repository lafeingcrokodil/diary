package recipe_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lafeingcrokodil/diary/internal/recipe"
)

func TestSortByName(t *testing.T) {
	t.Parallel()

	recipes := []recipe.Recipe{
		{ID: "b", Name: "soup"},
		{ID: "c", Name: "Bread"},
		{ID: "a", Name: "soup"},
		{ID: "d", Name: "apple pie"},
	}

	recipe.SortByName(recipes)

	ids := make([]string, 0, len(recipes))
	for _, item := range recipes {
		ids = append(ids, item.ID)
	}

	// Names are compared ignoring case, and equal names are ordered by ID.
	assert.Equal(t, []string{"d", "c", "a", "b"}, ids)
}
