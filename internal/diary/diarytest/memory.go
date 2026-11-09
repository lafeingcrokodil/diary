package diarytest

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"github.com/lafeingcrokodil/diary/internal/diary"
)

// MemoryStore is a diary.Store that keeps days in memory, for tests that need
// a store but aren't about how days are stored. It's safe for concurrent use.
type MemoryStore struct {
	mu   sync.Mutex
	days map[key]diary.Day
}

// key identifies one of a user's days.
type key struct {
	owner, date string
}

// NewMemoryStore returns an empty MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{days: map[key]diary.Day{}}
}

var _ diary.Store = (*MemoryStore)(nil)

// Get implements diary.Store.
func (s *MemoryStore) Get(_ context.Context, owner, date string) (diary.Day, error) {
	if err := diary.ValidateDate(date); err != nil {
		return diary.Day{}, fmt.Errorf("getting day: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	day, ok := s.days[key{owner, date}]
	if !ok {
		return diary.Day{Owner: owner, Date: date}, nil
	}

	return clone(day), nil
}

// Put implements diary.Store.
func (s *MemoryStore) Put(_ context.Context, owner string, day diary.Day) (diary.Day, error) {
	if err := diary.ValidateDate(day.Date); err != nil {
		return diary.Day{}, fmt.Errorf("storing day: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	day.Owner = owner
	day.Entries = diary.SortedEntries(day.Entries)

	if day.IsEmpty() {
		delete(s.days, key{owner, day.Date})
	} else {
		// Only the stored copy needs its own values: the returned day shares
		// them with the caller's input, which the store doesn't keep.
		s.days[key{owner, day.Date}] = clone(day)
	}

	return day, nil
}

// clone returns a copy of day that shares nothing with it, so that changing
// one doesn't change the other.
func clone(day diary.Day) diary.Day {
	day.WeightKg = clonePointer(day.WeightKg)
	day.KcalExpended = clonePointer(day.KcalExpended)
	day.Entries = slices.Clone(day.Entries)

	return day
}

func clonePointer(value *float64) *float64 {
	if value == nil {
		return nil
	}

	copied := *value

	return &copied
}
