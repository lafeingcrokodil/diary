// Package nutrient defines the nutrients that users track, and how they're
// stored.
package nutrient

import (
	"errors"
	"fmt"
)

// Unit is what a nutrient's amounts are measured in.
type Unit string

// The units that nutrients can be measured in.
const (
	Grams      Unit = "g"
	Milligrams Unit = "mg"
	Micrograms Unit = "µg"
)

// Nutrient is something in foods that a user tracks, such as protein.
type Nutrient struct {
	// ID identifies the nutrient. The store assigns it when the nutrient is
	// created.
	ID string
	// Owner is the ID of the user who the nutrient belongs to.
	Owner string
	Name  string
	Unit  Unit
	// KcalPerGram is the energy that a gram of the nutrient provides. It's
	// zero for nutrients that provide no energy, or whose energy isn't known.
	KcalPerGram float64
	// Archived nutrients are kept, along with their amounts in foods, but are
	// hidden until they're unarchived.
	Archived bool
}

// ErrInvalid is wrapped by the errors that Validate returns.
var ErrInvalid = errors.New("invalid nutrient")

// Validate checks the fields that users enter, and returns an error that wraps
// ErrInvalid if any of them is invalid.
func (n Nutrient) Validate() error {
	if n.Name == "" {
		return fmt.Errorf("%w: the name is required", ErrInvalid)
	}

	switch n.Unit {
	case Grams, Milligrams, Micrograms:
	default:
		return fmt.Errorf("%w: the unit must be %q, %q or %q", ErrInvalid, Grams, Milligrams, Micrograms)
	}

	if n.KcalPerGram < 0 {
		return fmt.Errorf("%w: the energy can't be negative", ErrInvalid)
	}

	return nil
}

// The energy per gram of the default nutrients, as used for food labels in the
// EU (Regulation (EU) No 1169/2011, Annex XIV).
const (
	kcalPerGramOfFat          = 9
	kcalPerGramOfCarbohydrate = 4
	kcalPerGramOfFibre        = 2
	kcalPerGramOfProtein      = 4
)

// Defaults returns the nutrients that users can start with, without IDs or
// owners. Carbohydrates don't include fibre, as on EU food labels.
func Defaults() []Nutrient {
	return []Nutrient{
		{Name: "Fat", Unit: Grams, KcalPerGram: kcalPerGramOfFat},
		{Name: "Carbohydrates", Unit: Grams, KcalPerGram: kcalPerGramOfCarbohydrate},
		{Name: "Fibre", Unit: Grams, KcalPerGram: kcalPerGramOfFibre},
		{Name: "Protein", Unit: Grams, KcalPerGram: kcalPerGramOfProtein},
	}
}
