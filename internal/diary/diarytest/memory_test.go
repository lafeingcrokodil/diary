package diarytest_test

import (
	"testing"

	"github.com/lafeingcrokodil/diary/internal/diary"
	"github.com/lafeingcrokodil/diary/internal/diary/diarytest"
)

func TestMemoryStore(t *testing.T) {
	t.Parallel()

	diarytest.CheckStore(t, func(*testing.T) diary.Store { return diarytest.NewMemoryStore() })
}
