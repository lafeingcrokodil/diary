package recipe

import (
	"cmp"
	"context"
	"errors"
	"slices"
	"strings"
)

// ErrNotFound is returned when a recipe doesn't exist, or belongs to another
// user.
var ErrNotFound = errors.New("recipe not found")

// ErrNotArchived is returned when deleting a recipe that isn't archived.
var ErrNotArchived = errors.New("recipe not archived")

// Store keeps users' recipes. Every method only acts on the given owner's
// recipes, so that users can't see or change each other's recipes.
type Store interface {
	// Create stores a new recipe for owner, and returns it with its ID and
	// owner set. Any ID or owner already set on recipe is ignored.
	Create(ctx context.Context, owner string, recipe Recipe) (Recipe, error)
	// Get returns one of owner's recipes, or ErrNotFound.
	Get(ctx context.Context, owner, id string) (Recipe, error)
	// List returns all of owner's recipes, including archived ones, sorted by
	// name.
	List(ctx context.Context, owner string) ([]Recipe, error)
	// Update replaces the recipe with recipe.ID, and returns it as stored. It
	// returns ErrNotFound if owner has no such recipe. The owner can't be
	// changed: any owner set on recipe is ignored.
	Update(ctx context.Context, owner string, recipe Recipe) (Recipe, error)
	// Delete removes one of owner's recipes. It returns ErrNotFound if owner
	// has no such recipe, and ErrNotArchived if the recipe isn't archived, so
	// that deleting one is always a second step after archiving it.
	Delete(ctx context.Context, owner, id string) error
}

// SortByName sorts recipes by name, ignoring case, as Store.List returns them.
// Recipes with the same name are sorted by ID, so that the order is stable.
func SortByName(recipes []Recipe) {
	slices.SortFunc(recipes, func(a, b Recipe) int {
		return cmp.Or(
			cmp.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)),
			cmp.Compare(a.ID, b.ID),
		)
	})
}
