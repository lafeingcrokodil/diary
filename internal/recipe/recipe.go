// Package recipe defines recipes, made from users' foods, and how they're
// stored.
package recipe

import (
	"errors"
	"fmt"
	"net/url"
)

// Ingredient is an amount of one of the owner's foods in a recipe.
type Ingredient struct {
	// FoodID is the ID of the food.
	FoodID string
	// Serving is the name of one of the food's servings, or empty if the
	// amount is in the food's unit.
	Serving string
	// Amount is the number of servings if Serving is set, and otherwise the
	// amount in the food's unit (e.g. in grams).
	Amount float64
	// Note is optional, e.g. "finely chopped".
	Note string
}

// Recipe is a list of ingredients and the steps to prepare them. Its yield is
// the whole recipe: amounts eaten are fractions of it.
type Recipe struct {
	// ID identifies the recipe. The store assigns it when the recipe is
	// created.
	ID string
	// Owner is the ID of the user who the recipe belongs to.
	Owner string
	Name  string
	// SourceURL is optional: where the recipe comes from.
	SourceURL   string
	Ingredients []Ingredient
	Steps       []string
	// Notes are optional, free-form text.
	Notes string
	// Archived recipes are kept, but are hidden until they're unarchived.
	Archived bool
}

// ErrInvalid is wrapped by the errors that Validate returns.
var ErrInvalid = errors.New("invalid recipe")

// Validate checks the fields that users enter, and returns an error that wraps
// ErrInvalid if any of them is invalid. It doesn't check that the ingredients'
// foods and servings exist, since that needs the owner's foods.
func (r Recipe) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("%w: the name is required", ErrInvalid)
	}

	if err := validateSourceURL(r.SourceURL); err != nil {
		return err
	}

	for index, ingredient := range r.Ingredients {
		if ingredient.FoodID == "" {
			return fmt.Errorf("%w: ingredient %d needs a food", ErrInvalid, index+1)
		}

		if ingredient.Amount <= 0 {
			return fmt.Errorf("%w: ingredient %d needs a positive amount", ErrInvalid, index+1)
		}
	}

	for i, step := range r.Steps {
		if step == "" {
			return fmt.Errorf("%w: step %d is empty", ErrInvalid, i+1)
		}
	}

	return nil
}

// validateSourceURL checks that sourceURL is empty, or an absolute HTTP or
// HTTPS URL, so that it can be shown as a link.
func validateSourceURL(sourceURL string) error {
	if sourceURL == "" {
		return nil
	}

	parsed, err := url.Parse(sourceURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return fmt.Errorf("%w: the source must be an http or https URL", ErrInvalid)
	}

	return nil
}
