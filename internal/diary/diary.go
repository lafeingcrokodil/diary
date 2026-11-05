// Package diary defines the days that users log, and how they're stored.
package diary

import (
	"errors"
	"fmt"
	"time"
)

// DateLayout is the format of a day's date, as a time.Parse layout.
const DateLayout = time.DateOnly

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
}

// IsEmpty reports whether nothing is logged for the day.
func (d Day) IsEmpty() bool {
	return d.WeightKg == nil && d.KcalExpended == nil
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
