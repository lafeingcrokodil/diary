package logging_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lafeingcrokodil/diary/internal/logging"
)

func TestNewHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		level        slog.Level
		wantSeverity string
	}{
		{slog.LevelDebug, "DEBUG"},
		{slog.LevelInfo, "INFO"},
		{slog.LevelWarn, "WARNING"},
		{slog.LevelWarn + 2, "WARNING"},
		{slog.LevelError, "ERROR"},
		{slog.LevelError + 4, "ERROR"},
	}

	for _, test := range tests {
		t.Run(test.level.String(), func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			logger := slog.New(logging.NewHandler(&buf, slog.LevelDebug))
			logger.Log(t.Context(), test.level, "hello", "err", "boom", slog.Group("req", "message", "nested"))

			var entry map[string]any

			require.NoError(t, json.Unmarshal(buf.Bytes(), &entry))

			assert.Subset(t, entry, map[string]any{
				"severity": test.wantSeverity,
				"message":  "hello",
				"err":      "boom",
				"req":      map[string]any{"message": "nested"},
			})
			assert.Contains(t, entry, "time")
			assert.NotContains(t, entry, "level")
			assert.NotContains(t, entry, "msg")
		})
	}
}

func TestNewHandlerLevel(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	logger := slog.New(logging.NewHandler(&buf, slog.LevelInfo))
	logger.DebugContext(t.Context(), "hidden")

	assert.Empty(t, buf.String(), "records below the handler's level should be dropped")
}
