package foodtest

import (
	"context"
	"crypto/rand"
	"maps"
	"slices"
	"sync"

	"github.com/lafeingcrokodil/diary/internal/food"
)

// MemoryStore is a food.Store that keeps foods in memory, for tests that need a
// store but aren't about how foods are stored. It's safe for concurrent use.
type MemoryStore struct {
	mu    sync.Mutex
	foods map[string]food.Food
}

// NewMemoryStore returns an empty MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{foods: map[string]food.Food{}}
}

var _ food.Store = (*MemoryStore)(nil)

// Create implements food.Store.
func (s *MemoryStore) Create(_ context.Context, owner string, item food.Food) (food.Food, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item.ID = rand.Text()
	item.Owner = owner
	// Only the stored copy needs its own servings: the returned food shares
	// them with the caller's input, which the store doesn't keep.
	s.foods[item.ID] = clone(item)

	return item, nil
}

// Get implements food.Store.
func (s *MemoryStore) Get(_ context.Context, owner, id string) (food.Food, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.foods[id]
	if !ok || item.Owner != owner {
		return food.Food{}, food.ErrNotFound
	}

	return clone(item), nil
}

// List implements food.Store.
func (s *MemoryStore) List(_ context.Context, owner string) ([]food.Food, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	foods := []food.Food{}

	for item := range maps.Values(s.foods) {
		if item.Owner == owner {
			foods = append(foods, clone(item))
		}
	}

	food.SortByName(foods)

	return foods, nil
}

// Update implements food.Store.
func (s *MemoryStore) Update(_ context.Context, owner string, item food.Food) (food.Food, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.foods[item.ID]
	if !ok || existing.Owner != owner {
		return food.Food{}, food.ErrNotFound
	}

	item.Owner = owner
	s.foods[item.ID] = clone(item)

	return item, nil
}

// Delete implements food.Store.
func (s *MemoryStore) Delete(_ context.Context, owner, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.foods[id]
	if !ok || item.Owner != owner {
		return food.ErrNotFound
	}

	delete(s.foods, id)

	return nil
}

// clone copies item's servings, so that the store and its callers don't share
// them.
func clone(item food.Food) food.Food {
	item.Servings = slices.Clone(item.Servings)

	return item
}
