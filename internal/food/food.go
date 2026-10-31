// Package food defines foods, with their nutritional information, and how
// they're stored.
package food

import (
	"errors"
	"fmt"
)

// Unit is what a food's amounts are measured in.
type Unit string

// The units that foods can be measured in. Converting between them would need
// each food's density, so a food sticks to one.
const (
	Grams       Unit = "g"
	Millilitres Unit = "ml"
)

// Serving is a named amount of a food, e.g. "1 slice", so that it can be
// logged without weighing it.
type Serving struct {
	Name string
	// Amount is the serving's size in the food's unit.
	Amount float64
}

// Food is something that can be eaten, with its nutritional information per
// 100 of its unit (e.g. per 100 g).
type Food struct {
	// ID identifies the food. The store assigns it when the food is created.
	ID string
	// Owner is the ID of the user who the food belongs to.
	Owner string
	Name  string
	// Brand is optional.
	Brand      string
	Unit       Unit
	KcalPer100 float64
	Servings   []Serving
}

// ErrInvalid is wrapped by the errors that Validate returns.
var ErrInvalid = errors.New("invalid food")

// Validate checks the fields that users enter, and returns an error that wraps
// ErrInvalid if any of them is invalid.
func (f Food) Validate() error {
	if f.Name == "" {
		return fmt.Errorf("%w: the name is required", ErrInvalid)
	}

	if f.Unit != Grams && f.Unit != Millilitres {
		return fmt.Errorf("%w: the unit must be %q or %q", ErrInvalid, Grams, Millilitres)
	}

	if f.KcalPer100 < 0 {
		return fmt.Errorf("%w: the energy can't be negative", ErrInvalid)
	}

	names := make(map[string]bool, len(f.Servings))

	for _, serving := range f.Servings {
		if serving.Name == "" {
			return fmt.Errorf("%w: every serving needs a name", ErrInvalid)
		}

		if serving.Amount <= 0 {
			return fmt.Errorf("%w: serving %q needs a positive amount", ErrInvalid, serving.Name)
		}

		if names[serving.Name] {
			return fmt.Errorf("%w: there's more than one serving called %q", ErrInvalid, serving.Name)
		}

		names[serving.Name] = true
	}

	return nil
}
