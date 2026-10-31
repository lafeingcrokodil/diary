package food

import (
	"context"
	"crypto/rand"
	"maps"
	"slices"
	"sync"
)

// MemoryStore is a Store that keeps foods in memory, which is enough for
// tests. It's safe for concurrent use.
type MemoryStore struct {
	mu    sync.Mutex
	foods map[string]Food
}

// NewMemoryStore returns an empty MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{foods: map[string]Food{}}
}

var _ Store = (*MemoryStore)(nil)

// Create implements Store.
func (s *MemoryStore) Create(_ context.Context, owner string, food Food) (Food, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	food.ID = rand.Text()
	food.Owner = owner
	// Only the stored copy needs its own servings: the returned food shares
	// them with the caller's input, which the store doesn't keep.
	s.foods[food.ID] = clone(food)

	return food, nil
}

// Get implements Store.
func (s *MemoryStore) Get(_ context.Context, owner, id string) (Food, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	food, ok := s.foods[id]
	if !ok || food.Owner != owner {
		return Food{}, ErrNotFound
	}

	return clone(food), nil
}

// List implements Store.
func (s *MemoryStore) List(_ context.Context, owner string) ([]Food, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	foods := []Food{}

	for food := range maps.Values(s.foods) {
		if food.Owner == owner {
			foods = append(foods, clone(food))
		}
	}

	SortByName(foods)

	return foods, nil
}

// Update implements Store.
func (s *MemoryStore) Update(_ context.Context, owner string, food Food) (Food, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.foods[food.ID]
	if !ok || existing.Owner != owner {
		return Food{}, ErrNotFound
	}

	food.Owner = owner
	s.foods[food.ID] = clone(food)

	return food, nil
}

// Delete implements Store.
func (s *MemoryStore) Delete(_ context.Context, owner, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	food, ok := s.foods[id]
	if !ok || food.Owner != owner {
		return ErrNotFound
	}

	delete(s.foods, id)

	return nil
}

// clone copies food's servings, so that the store and its callers don't share
// them.
func clone(food Food) Food {
	food.Servings = slices.Clone(food.Servings)

	return food
}
