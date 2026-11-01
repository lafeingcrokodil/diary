package nutrienttest

import (
	"context"
	"crypto/rand"
	"maps"
	"sync"

	"github.com/lafeingcrokodil/diary/internal/nutrient"
)

// MemoryStore is a nutrient.Store that keeps nutrients in memory, for tests
// that need a store but aren't about how nutrients are stored. It's safe for
// concurrent use.
type MemoryStore struct {
	mu        sync.Mutex
	nutrients map[string]nutrient.Nutrient
}

// NewMemoryStore returns an empty MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{nutrients: map[string]nutrient.Nutrient{}}
}

var _ nutrient.Store = (*MemoryStore)(nil)

// Create implements nutrient.Store.
func (s *MemoryStore) Create(_ context.Context, owner string, item nutrient.Nutrient) (nutrient.Nutrient, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item.ID = rand.Text()
	item.Owner = owner
	s.nutrients[item.ID] = item

	return item, nil
}

// Get implements nutrient.Store.
func (s *MemoryStore) Get(_ context.Context, owner, id string) (nutrient.Nutrient, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.nutrients[id]
	if !ok || item.Owner != owner {
		return nutrient.Nutrient{}, nutrient.ErrNotFound
	}

	return item, nil
}

// List implements nutrient.Store.
func (s *MemoryStore) List(_ context.Context, owner string) ([]nutrient.Nutrient, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	nutrients := []nutrient.Nutrient{}

	for item := range maps.Values(s.nutrients) {
		if item.Owner == owner {
			nutrients = append(nutrients, item)
		}
	}

	nutrient.SortByName(nutrients)

	return nutrients, nil
}

// Update implements nutrient.Store.
func (s *MemoryStore) Update(_ context.Context, owner string, item nutrient.Nutrient) (nutrient.Nutrient, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.nutrients[item.ID]
	if !ok || existing.Owner != owner {
		return nutrient.Nutrient{}, nutrient.ErrNotFound
	}

	item.Owner = owner
	s.nutrients[item.ID] = item

	return item, nil
}

// Delete implements nutrient.Store.
func (s *MemoryStore) Delete(_ context.Context, owner, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.nutrients[id]
	if !ok || item.Owner != owner {
		return nutrient.ErrNotFound
	}

	delete(s.nutrients, id)

	return nil
}
