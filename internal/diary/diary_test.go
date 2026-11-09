package diary_test

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lafeingcrokodil/diary/internal/diary"
	"github.com/lafeingcrokodil/diary/internal/diary/diarytest"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	zero, negative := 0.0, -1.0

	tests := []struct {
		name    string
		change  func(d *diary.Day)
		wantErr bool
	}{
		{"valid", func(*diary.Day) {}, false},
		{"empty", func(d *diary.Day) { *d = diary.Day{Date: d.Date} }, false},
		{"no weight", func(d *diary.Day) { d.WeightKg = nil }, false},
		{"no energy expended", func(d *diary.Day) { d.KcalExpended = nil }, false},
		{"no energy expended at all", func(d *diary.Day) { d.KcalExpended = &zero }, false},
		{"leap day", func(d *diary.Day) { d.Date = "2028-02-29" }, false},
		{"no entries", func(d *diary.Day) { d.Entries = nil }, false},
		{"entries out of order", func(d *diary.Day) { slices.Reverse(d.Entries) }, false},
		{"entry at midnight", func(d *diary.Day) { d.Entries[0].Time = "00:00" }, false},
		{"entry just before midnight", func(d *diary.Day) { d.Entries[0].Time = "23:59" }, false},
		{"fraction of a serving", func(d *diary.Day) { d.Entries[2].Amount = 0.5 }, false},
		{"more than a whole recipe", func(d *diary.Day) { d.Entries[1].Amount = 1.5 }, false},
		{"no date", func(d *diary.Day) { d.Date = "" }, true},
		{"date with a time", func(d *diary.Day) { d.Date = "2026-11-05T12:00:00Z" }, true},
		{"date in another format", func(d *diary.Day) { d.Date = "05/11/2026" }, true},
		{"date without leading zeros", func(d *diary.Day) { d.Date = "2026-1-5" }, true},
		{"date that doesn't exist", func(d *diary.Day) { d.Date = "2026-02-29" }, true},
		{"date with a slash", func(d *diary.Day) { d.Date = "2026/11/05" }, true},
		{"zero weight", func(d *diary.Day) { d.WeightKg = &zero }, true},
		{"negative weight", func(d *diary.Day) { d.WeightKg = &negative }, true},
		{"negative energy expended", func(d *diary.Day) { d.KcalExpended = &negative }, true},
		{"entry without a time", func(d *diary.Day) { d.Entries[0].Time = "" }, true},
		{"entry with seconds", func(d *diary.Day) { d.Entries[0].Time = "08:15:00" }, true},
		{"entry without a leading zero", func(d *diary.Day) { d.Entries[0].Time = "8:15" }, true},
		{"entry at 24:00", func(d *diary.Day) { d.Entries[0].Time = "24:00" }, true},
		{"entry in 12-hour time", func(d *diary.Day) { d.Entries[0].Time = "8:15am" }, true},
		{"entry without a food or recipe", func(d *diary.Day) { d.Entries[0].FoodID = "" }, true},
		{"entry with a food and a recipe", func(d *diary.Day) { d.Entries[0].RecipeID = "soup" }, true},
		{"recipe entry with a serving", func(d *diary.Day) { d.Entries[1].Serving = "bowl" }, true},
		{"entry without an amount", func(d *diary.Day) { d.Entries[0].Amount = 0 }, true},
		{"entry with a negative amount", func(d *diary.Day) { d.Entries[1].Amount = -0.5 }, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			d := diarytest.Monday()
			test.change(&d)

			err := d.Validate()
			if test.wantErr {
				require.ErrorIs(t, err, diary.ErrInvalid)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsEmpty(t *testing.T) {
	t.Parallel()

	weight, kcal := 70.0, 0.0

	assert.True(t, diary.Day{Owner: "alice", Date: "2026-11-02"}.IsEmpty())
	assert.False(t, diary.Day{WeightKg: &weight}.IsEmpty())
	// Logging that no energy was expended is still logging something.
	assert.False(t, diary.Day{KcalExpended: &kcal}.IsEmpty())
	assert.False(t, diary.Day{Entries: diarytest.Monday().Entries}.IsEmpty())
	assert.True(t, diary.Day{Entries: []diary.Entry{}}.IsEmpty())
}

func TestSortedEntries(t *testing.T) {
	t.Parallel()

	entries := []diary.Entry{
		{Time: "19:00", FoodID: "rice", Amount: 1},
		{Time: "07:30", FoodID: "oats", Amount: 2},
		{Time: "19:00", FoodID: "beans", Amount: 3},
		{Time: "12:00", RecipeID: "soup", Amount: 4},
	}
	original := slices.Clone(entries)

	sorted := diary.SortedEntries(entries)

	amounts := make([]float64, 0, len(sorted))
	for _, entry := range sorted {
		amounts = append(amounts, entry.Amount)
	}

	// Sorted by time, with the two entries at 19:00 in their original order.
	assert.Equal(t, []float64{2, 4, 1, 3}, amounts)
	assert.Equal(t, original, entries, "the input is left alone")

	assert.Nil(t, diary.SortedEntries(nil))
	assert.Nil(t, diary.SortedEntries([]diary.Entry{}))
}
