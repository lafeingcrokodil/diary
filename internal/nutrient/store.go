package nutrient

import (
	"cmp"
	"context"
	"errors"
	"slices"
	"strings"
)

// ErrNotFound is returned when a nutrient doesn't exist, or belongs to another
// user.
var ErrNotFound = errors.New("nutrient not found")

// ErrNotArchived is returned when deleting a nutrient that isn't archived.
var ErrNotArchived = errors.New("nutrient not archived")

// Store keeps users' nutrients. Every method only acts on the given owner's
// nutrients, so that users can't see or change each other's nutrients.
type Store interface {
	// Create stores a new nutrient for owner, and returns it with its ID and
	// owner set. Any ID or owner already set on nutrient is ignored.
	Create(ctx context.Context, owner string, nutrient Nutrient) (Nutrient, error)
	// Get returns one of owner's nutrients, or ErrNotFound.
	Get(ctx context.Context, owner, id string) (Nutrient, error)
	// List returns all of owner's nutrients, including archived ones, sorted
	// by name.
	List(ctx context.Context, owner string) ([]Nutrient, error)
	// Update replaces the nutrient with nutrient.ID, and returns it as stored.
	// It returns ErrNotFound if owner has no such nutrient. The owner can't be
	// changed: any owner set on nutrient is ignored.
	Update(ctx context.Context, owner string, nutrient Nutrient) (Nutrient, error)
	// Delete removes one of owner's nutrients, along with its amounts in
	// owner's foods, all at once. It returns ErrNotFound if owner has no such
	// nutrient, and ErrNotArchived if the nutrient isn't archived, so that
	// deleting one is always a second step after archiving it.
	Delete(ctx context.Context, owner, id string) error
}

// SortByName sorts nutrients by name, ignoring case, as Store.List returns
// them. Nutrients with the same name are sorted by ID, so that the order is
// stable.
func SortByName(nutrients []Nutrient) {
	slices.SortFunc(nutrients, func(a, b Nutrient) int {
		return cmp.Or(
			cmp.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)),
			cmp.Compare(a.ID, b.ID),
		)
	})
}
