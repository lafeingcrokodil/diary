package recipe_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lafeingcrokodil/diary/internal/recipe"
	"github.com/lafeingcrokodil/diary/internal/recipe/recipetest"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		change  func(r *recipe.Recipe)
		wantErr bool
	}{
		{"valid", func(*recipe.Recipe) {}, false},
		{"archived", func(r *recipe.Recipe) { r.Archived = true }, false},
		{"no source", func(r *recipe.Recipe) { r.SourceURL = "" }, false},
		{"http source", func(r *recipe.Recipe) { r.SourceURL = "http://example.com/soup" }, false},
		{"no notes", func(r *recipe.Recipe) { r.Notes = "" }, false},
		{"no ingredients", func(r *recipe.Recipe) { r.Ingredients = nil }, false},
		{"no steps", func(r *recipe.Recipe) { r.Steps = nil }, false},
		{"ingredient without a note", func(r *recipe.Recipe) { r.Ingredients[0].Note = "" }, false},
		{"same food twice", func(r *recipe.Recipe) { r.Ingredients[1].FoodID = r.Ingredients[0].FoodID }, false},
		{"fraction of a serving", func(r *recipe.Recipe) { r.Ingredients[1].Amount = 0.5 }, false},
		{"no name", func(r *recipe.Recipe) { r.Name = "" }, true},
		{"relative source", func(r *recipe.Recipe) { r.SourceURL = "/soup" }, true},
		{"source without a host", func(r *recipe.Recipe) { r.SourceURL = "https:///soup" }, true},
		{"source with another scheme", func(r *recipe.Recipe) { r.SourceURL = "javascript:alert(1)" }, true},
		{"unparsable source", func(r *recipe.Recipe) { r.SourceURL = "https://example.com/%zz" }, true},
		{"ingredient without a food", func(r *recipe.Recipe) { r.Ingredients[0].FoodID = "" }, true},
		{"ingredient without an amount", func(r *recipe.Recipe) { r.Ingredients[0].Amount = 0 }, true},
		{"ingredient with a negative amount", func(r *recipe.Recipe) { r.Ingredients[1].Amount = -1 }, true},
		{"empty step", func(r *recipe.Recipe) { r.Steps[1] = "" }, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			r := recipetest.Soup()
			test.change(&r)

			err := r.Validate()
			if test.wantErr {
				require.ErrorIs(t, err, recipe.ErrInvalid)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
