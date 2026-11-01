package nutrient_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lafeingcrokodil/diary/internal/nutrient"
)

func TestSortByName(t *testing.T) {
	t.Parallel()

	nutrients := []nutrient.Nutrient{
		{ID: "b", Name: "protein"},
		{ID: "c", Name: "Fibre"},
		{ID: "a", Name: "protein"},
		{ID: "d", Name: "carbohydrates"},
	}

	nutrient.SortByName(nutrients)

	ids := make([]string, 0, len(nutrients))
	for _, n := range nutrients {
		ids = append(ids, n.ID)
	}

	// Names are compared ignoring case, and equal names are ordered by ID.
	assert.Equal(t, []string{"d", "c", "a", "b"}, ids)
}
