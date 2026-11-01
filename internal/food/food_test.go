package food_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lafeingcrokodil/diary/internal/food"
	"github.com/lafeingcrokodil/diary/internal/food/foodtest"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		change  func(f *food.Food)
		wantErr bool
	}{
		{"valid", func(*food.Food) {}, false},
		{"millilitres", func(f *food.Food) { f.Unit = food.Millilitres }, false},
		{"no brand", func(f *food.Food) { f.Brand = "" }, false},
		{"no servings", func(f *food.Food) { f.Servings = nil }, false},
		{"no energy", func(f *food.Food) { f.KcalPer100 = 0 }, false},
		{"no name", func(f *food.Food) { f.Name = "" }, true},
		{"no unit", func(f *food.Food) { f.Unit = "" }, true},
		{"unknown unit", func(f *food.Food) { f.Unit = "oz" }, true},
		{"negative energy", func(f *food.Food) { f.KcalPer100 = -1 }, true},
		{"no nutrients", func(f *food.Food) { f.Nutrients = nil }, false},
		{"no amount of a nutrient", func(f *food.Food) { f.Nutrients["protein"] = 0 }, false},
		{"negative amount of a nutrient", func(f *food.Food) { f.Nutrients["protein"] = -1 }, true},
		{"nutrient without an ID", func(f *food.Food) { f.Nutrients[""] = 1 }, true},
		{"serving without a name", func(f *food.Food) { f.Servings[0].Name = "" }, true},
		{"serving without an amount", func(f *food.Food) { f.Servings[0].Amount = 0 }, true},
		{"serving with a negative amount", func(f *food.Food) { f.Servings[0].Amount = -40 }, true},
		{"servings with the same name", func(f *food.Food) { f.Servings[1].Name = "slice" }, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			f := foodtest.Bread()
			test.change(&f)

			err := f.Validate()
			if test.wantErr {
				require.ErrorIs(t, err, food.ErrInvalid)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
