package food

import (
	"cmp"
	"context"
	"errors"
	"slices"
	"strings"
)

// ErrNotFound is returned when a food doesn't exist, or belongs to another
// user.
var ErrNotFound = errors.New("food not found")

// Store keeps users' foods. Every method only acts on the given owner's foods,
// so that users can't see or change each other's foods.
type Store interface {
	// Create stores a new food for owner, and returns it with its ID and owner
	// set. Any ID or owner already set on food is ignored.
	Create(ctx context.Context, owner string, food Food) (Food, error)
	// Get returns one of owner's foods, or ErrNotFound.
	Get(ctx context.Context, owner, id string) (Food, error)
	// List returns all of owner's foods, sorted by name.
	List(ctx context.Context, owner string) ([]Food, error)
	// Update replaces the food with food.ID, and returns it as stored. It
	// returns ErrNotFound if owner has no such food. The owner can't be
	// changed: any owner set on food is ignored.
	Update(ctx context.Context, owner string, food Food) (Food, error)
	// Delete removes one of owner's foods, or returns ErrNotFound.
	Delete(ctx context.Context, owner, id string) error
}

// SortByName sorts foods by name, ignoring case, as Store.List returns them.
// Foods with the same name are sorted by ID, so that the order is stable.
func SortByName(foods []Food) {
	slices.SortFunc(foods, func(a, b Food) int {
		return cmp.Or(
			cmp.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)),
			cmp.Compare(a.ID, b.ID),
		)
	})
}
