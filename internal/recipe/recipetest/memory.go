package recipetest

import (
	"context"
	"crypto/rand"
	"maps"
	"slices"
	"sync"

	"github.com/lafeingcrokodil/diary/internal/recipe"
)

// MemoryStore is a recipe.Store that keeps recipes in memory, for tests that
// need a store but aren't about how recipes are stored. It's safe for
// concurrent use.
type MemoryStore struct {
	mu      sync.Mutex
	recipes map[string]recipe.Recipe
}

// NewMemoryStore returns an empty MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{recipes: map[string]recipe.Recipe{}}
}

var _ recipe.Store = (*MemoryStore)(nil)

// Create implements recipe.Store.
func (s *MemoryStore) Create(_ context.Context, owner string, item recipe.Recipe) (recipe.Recipe, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item.ID = rand.Text()
	item.Owner = owner
	// Only the stored copy needs its own ingredients and steps: the returned
	// recipe shares them with the caller's input, which the store doesn't
	// keep.
	s.recipes[item.ID] = clone(item)

	return item, nil
}

// Get implements recipe.Store.
func (s *MemoryStore) Get(_ context.Context, owner, id string) (recipe.Recipe, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.recipes[id]
	if !ok || item.Owner != owner {
		return recipe.Recipe{}, recipe.ErrNotFound
	}

	return clone(item), nil
}

// List implements recipe.Store.
func (s *MemoryStore) List(_ context.Context, owner string) ([]recipe.Recipe, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	recipes := []recipe.Recipe{}

	for item := range maps.Values(s.recipes) {
		if item.Owner == owner {
			recipes = append(recipes, clone(item))
		}
	}

	recipe.SortByName(recipes)

	return recipes, nil
}

// Update implements recipe.Store.
func (s *MemoryStore) Update(_ context.Context, owner string, item recipe.Recipe) (recipe.Recipe, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.recipes[item.ID]
	if !ok || existing.Owner != owner {
		return recipe.Recipe{}, recipe.ErrNotFound
	}

	item.Owner = owner
	s.recipes[item.ID] = clone(item)

	return item, nil
}

// Delete implements recipe.Store.
func (s *MemoryStore) Delete(_ context.Context, owner, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.recipes[id]
	if !ok || item.Owner != owner {
		return recipe.ErrNotFound
	}

	if !item.Archived {
		return recipe.ErrNotArchived
	}

	delete(s.recipes, id)

	return nil
}

// clone copies item's ingredients and steps, so that the store and its
// callers don't share them.
func clone(item recipe.Recipe) recipe.Recipe {
	item.Ingredients = slices.Clone(item.Ingredients)
	item.Steps = slices.Clone(item.Steps)

	return item
}
