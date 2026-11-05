// Package diarytest checks that implementations of diary.Store behave the same
// way, so that each implementation's tests can run the same checks.
package diarytest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lafeingcrokodil/diary/internal/diary"
)

const (
	alice = "alice"
	bob   = "bob"
)

// Monday returns a valid day with everything logged, which tests can change as
// they need.
func Monday() diary.Day {
	weight, kcal := 70.5, 2300.0

	return diary.Day{
		Date:         "2026-11-02",
		WeightKg:     &weight,
		KcalExpended: &kcal,
	}
}

// CheckStore checks that a Store implementation behaves as the interface
// describes. newStore must return an empty store each time it's called. It
// gets the test that the store is for, so that it can fail that test if it
// can't create a store.
func CheckStore(t *testing.T, newStore func(t *testing.T) diary.Store) {
	t.Helper()

	t.Run("put and get", func(t *testing.T) {
		t.Parallel()

		store := newStore(t)

		input := Monday()
		input.Owner = bob

		stored, err := store.Put(t.Context(), alice, input)
		require.NoError(t, err)

		want := Monday()
		want.Owner = alice
		assert.Equal(t, want, stored)

		got, err := store.Get(t.Context(), alice, want.Date)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("a day can have only some things logged", func(t *testing.T) {
		t.Parallel()

		store := newStore(t)

		input := Monday()
		input.WeightKg = nil

		_, err := store.Put(t.Context(), alice, input)
		require.NoError(t, err)

		got, err := store.Get(t.Context(), alice, input.Date)
		require.NoError(t, err)
		assert.Nil(t, got.WeightKg)
		assert.Equal(t, input.KcalExpended, got.KcalExpended)
	})

	t.Run("get a day that nothing is logged for", func(t *testing.T) {
		t.Parallel()

		got, err := newStore(t).Get(t.Context(), alice, "2026-11-02")
		require.NoError(t, err)
		assert.Equal(t, diary.Day{Owner: alice, Date: "2026-11-02"}, got)
	})

	t.Run("put replaces the day", func(t *testing.T) {
		t.Parallel()

		store := newStore(t)

		_, err := store.Put(t.Context(), alice, Monday())
		require.NoError(t, err)

		weight := 69.8
		changed := Monday()
		changed.WeightKg = &weight
		changed.KcalExpended = nil

		_, err = store.Put(t.Context(), alice, changed)
		require.NoError(t, err)

		got, err := store.Get(t.Context(), alice, changed.Date)
		require.NoError(t, err)

		changed.Owner = alice
		assert.Equal(t, changed, got)
	})

	t.Run("putting an empty day removes what was logged", func(t *testing.T) {
		t.Parallel()

		store := newStore(t)

		_, err := store.Put(t.Context(), alice, Monday())
		require.NoError(t, err)

		empty := diary.Day{Date: Monday().Date}

		stored, err := store.Put(t.Context(), alice, empty)
		require.NoError(t, err)

		empty.Owner = alice
		assert.Equal(t, empty, stored)

		got, err := store.Get(t.Context(), alice, empty.Date)
		require.NoError(t, err)
		assert.Equal(t, empty, got)

		// Putting a day that's already empty is fine too.
		_, err = store.Put(t.Context(), alice, empty)
		require.NoError(t, err)
	})

	t.Run("days are kept apart", func(t *testing.T) {
		t.Parallel()

		store := newStore(t)

		_, err := store.Put(t.Context(), alice, Monday())
		require.NoError(t, err)

		got, err := store.Get(t.Context(), alice, "2026-11-03")
		require.NoError(t, err)
		assert.True(t, got.IsEmpty())
	})

	t.Run("invalid dates are rejected", func(t *testing.T) {
		t.Parallel()

		store := newStore(t)

		for _, date := range []string{"", "/", "2026-11-02/", "days/2026-11-02", "2026-02-30"} {
			_, err := store.Get(t.Context(), alice, date)
			require.ErrorIs(t, err, diary.ErrInvalid, "Get(%q)", date)

			day := Monday()
			day.Date = date
			_, err = store.Put(t.Context(), alice, day)
			require.ErrorIs(t, err, diary.ErrInvalid, "Put(%q)", date)
		}
	})

	t.Run("users can't reach each other's days", func(t *testing.T) {
		t.Parallel()

		store := newStore(t)

		alicesDay, err := store.Put(t.Context(), alice, Monday())
		require.NoError(t, err)

		got, err := store.Get(t.Context(), bob, alicesDay.Date)
		require.NoError(t, err)
		assert.True(t, got.IsEmpty())

		// Bob logging the same date, or emptying it, leaves Alice's day alone.
		weight := 90.0
		bobsDay := Monday()
		bobsDay.WeightKg = &weight

		_, err = store.Put(t.Context(), bob, bobsDay)
		require.NoError(t, err)

		_, err = store.Put(t.Context(), bob, diary.Day{Date: alicesDay.Date})
		require.NoError(t, err)

		got, err = store.Get(t.Context(), alice, alicesDay.Date)
		require.NoError(t, err)
		assert.Equal(t, alicesDay, got)
	})
}
