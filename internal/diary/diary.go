// Package diary defines the days that users log, with what they ate, and how
// they're stored.
package diary

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"time"
)

// DateLayout is the format of a day's date, as a time.Parse layout.
const DateLayout = time.DateOnly

// TimeLayout is the format of an entry's time of day, as a time.Parse layout.
const TimeLayout = "15:04"

// Entry is something that the user ate: either an amount of one of their
// foods, or a fraction of one of their recipes.
type Entry struct {
	// Time is the local time of day, formatted with TimeLayout.
	Time string
	// FoodID is the ID of the food, or empty if the entry is a recipe.
	FoodID string
	// Serving is the name of one of the food's servings, or empty if the
	// amount is in the food's unit. Recipe entries don't have one.
	Serving string
	// RecipeID is the ID of the recipe, or empty if the entry is a food.
	RecipeID string
	// Amount is, for a food, the number of servings if Serving is set, and
	// otherwise the amount in the food's unit (e.g. in grams). For a recipe,
	// it's the fraction of the whole recipe (e.g. 0.5 for half of it).
	Amount float64
}

// Day is what a user logs for one calendar day.
type Day struct {
	// Owner is the ID of the user who the day belongs to.
	Owner string
	// Date is the calendar day, formatted with DateLayout. It's a date
	// without a time zone, so it means whichever day it is wherever the user
	// is.
	Date string
	// WeightKg is the user's body weight in kilograms, or nil if it isn't
	// logged.
	WeightKg *float64
	// KcalExpended is the energy that the user expended over the whole day,
	// or nil if it isn't logged.
	KcalExpended *float64
	// Entries are what the user ate, sorted by time (see SortedEntries).
	Entries []Entry
}

// IsEmpty reports whether nothing is logged for the day.
func (d Day) IsEmpty() bool {
	return d.WeightKg == nil && d.KcalExpended == nil && len(d.Entries) == 0
}

// SortedEntries returns a copy of entries sorted by time, in the order that
// days keep them. Entries with the same time keep their order. It returns nil
// if there are no entries, so that a day without entries has nil Entries,
// however they were given.
func SortedEntries(entries []Entry) []Entry {
	if len(entries) == 0 {
		return nil
	}

	sorted := slices.Clone(entries)
	slices.SortStableFunc(sorted, func(a, b Entry) int { return cmp.Compare(a.Time, b.Time) })

	return sorted
}

// ErrInvalid is wrapped by the errors that Validate and ValidateDate return.
var ErrInvalid = errors.New("invalid day")

// Validate checks the fields that users enter, and returns an error that wraps
// ErrInvalid if any of them is invalid.
func (d Day) Validate() error {
	if err := ValidateDate(d.Date); err != nil {
		return err
	}

	if d.WeightKg != nil && *d.WeightKg <= 0 {
		return fmt.Errorf("%w: the weight must be positive", ErrInvalid)
	}

	if d.KcalExpended != nil && *d.KcalExpended < 0 {
		return fmt.Errorf("%w: the energy expended can't be negative", ErrInvalid)
	}

	for index, entry := range d.Entries {
		if err := entry.validate(); err != nil {
			return fmt.Errorf("entry %d: %w", index+1, err)
		}
	}

	return nil
}

// validate checks an entry's fields, and returns an error that wraps
// ErrInvalid if any of them is invalid. It doesn't check that the entry's food
// or recipe exists, since that needs the owner's foods and recipes.
func (e Entry) validate() error {
	// Times are compared as text when sorting entries, which only works if
	// they're all in exactly the same format.
	parsed, err := time.Parse(TimeLayout, e.Time)
	if err != nil || parsed.Format(TimeLayout) != e.Time {
		return fmt.Errorf("%w: the time must be a time like 13:05", ErrInvalid)
	}

	if (e.FoodID == "") == (e.RecipeID == "") {
		return fmt.Errorf("%w: an entry must have either a food or a recipe", ErrInvalid)
	}

	if e.RecipeID != "" && e.Serving != "" {
		return fmt.Errorf("%w: a recipe entry can't have a serving", ErrInvalid)
	}

	if e.Amount <= 0 {
		return fmt.Errorf("%w: the amount must be positive", ErrInvalid)
	}

	return nil
}

// ValidateDate checks that date is a real calendar date formatted with
// DateLayout, and returns an error that wraps ErrInvalid if it isn't.
func ValidateDate(date string) error {
	// Parsing alone would accept some other forms of the same date, so the
	// date must also be what formatting it gives back.
	parsed, err := time.Parse(DateLayout, date)
	if err != nil || parsed.Format(DateLayout) != date {
		return fmt.Errorf("%w: the date must be a date like 2026-01-31", ErrInvalid)
	}

	return nil
}
