package diary

import "context"

// Store keeps users' days. Every method only acts on the given owner's days,
// so that users can't see or change each other's days.
type Store interface {
	// Get returns owner's day with the given date. A day that nothing is
	// logged for is returned empty, rather than as an error. It returns an
	// error that wraps ErrInvalid if the date isn't valid (see ValidateDate).
	Get(ctx context.Context, owner, date string) (Day, error)
	// Put replaces owner's day with day.Date, and returns it as stored.
	// Putting an empty day removes what was logged for it. The owner can't be
	// changed: any owner set on day is ignored. It returns an error that wraps
	// ErrInvalid if the date isn't valid (see ValidateDate).
	Put(ctx context.Context, owner string, day Day) (Day, error)
}
