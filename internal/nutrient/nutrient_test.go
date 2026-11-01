package nutrient_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lafeingcrokodil/diary/internal/nutrient"
	"github.com/lafeingcrokodil/diary/internal/nutrient/nutrienttest"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		change  func(n *nutrient.Nutrient)
		wantErr bool
	}{
		{"valid", func(*nutrient.Nutrient) {}, false},
		{"milligrams", func(n *nutrient.Nutrient) { n.Unit = nutrient.Milligrams }, false},
		{"micrograms", func(n *nutrient.Nutrient) { n.Unit = nutrient.Micrograms }, false},
		{"no energy", func(n *nutrient.Nutrient) { n.KcalPerGram = 0 }, false},
		{"archived", func(n *nutrient.Nutrient) { n.Archived = true }, false},
		{"no name", func(n *nutrient.Nutrient) { n.Name = "" }, true},
		{"no unit", func(n *nutrient.Nutrient) { n.Unit = "" }, true},
		{"unknown unit", func(n *nutrient.Nutrient) { n.Unit = "oz" }, true},
		{"negative energy", func(n *nutrient.Nutrient) { n.KcalPerGram = -1 }, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			n := nutrienttest.Protein()
			test.change(&n)

			err := n.Validate()
			if test.wantErr {
				require.ErrorIs(t, err, nutrient.ErrInvalid)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDefaults(t *testing.T) {
	t.Parallel()

	defaults := nutrient.Defaults()

	names := make([]string, 0, len(defaults))
	for _, n := range defaults {
		require.NoError(t, n.Validate(), n.Name)
		assert.Positive(t, n.KcalPerGram, n.Name)
		names = append(names, n.Name)
	}

	assert.Equal(t, []string{"Fat", "Carbohydrates", "Fibre", "Protein"}, names)

	// Each call returns a new slice, so callers can't change later results.
	defaults[0].Name = "Changed"
	assert.Equal(t, "Fat", nutrient.Defaults()[0].Name)
}
