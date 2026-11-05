package diary_test

import (
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
		{"no date", func(d *diary.Day) { d.Date = "" }, true},
		{"date with a time", func(d *diary.Day) { d.Date = "2026-11-05T12:00:00Z" }, true},
		{"date in another format", func(d *diary.Day) { d.Date = "05/11/2026" }, true},
		{"date without leading zeros", func(d *diary.Day) { d.Date = "2026-1-5" }, true},
		{"date that doesn't exist", func(d *diary.Day) { d.Date = "2026-02-29" }, true},
		{"date with a slash", func(d *diary.Day) { d.Date = "2026/11/05" }, true},
		{"zero weight", func(d *diary.Day) { d.WeightKg = &zero }, true},
		{"negative weight", func(d *diary.Day) { d.WeightKg = &negative }, true},
		{"negative energy expended", func(d *diary.Day) { d.KcalExpended = &negative }, true},
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
}
