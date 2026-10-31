package logging_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"reflect"
	"testing"

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

			err := json.Unmarshal(buf.Bytes(), &entry)
			if err != nil {
				t.Fatalf("parsing %q: %v", buf.String(), err)
			}

			want := map[string]any{
				"severity": test.wantSeverity,
				"message":  "hello",
				"err":      "boom",
				"req":      map[string]any{"message": "nested"},
			}

			for key, wantValue := range want {
				if !reflect.DeepEqual(entry[key], wantValue) {
					t.Errorf("%s = %v, want %v", key, entry[key], wantValue)
				}
			}

			for _, key := range []string{"level", "msg"} {
				if _, ok := entry[key]; ok {
					t.Errorf("unexpected %q field in %q", key, buf.String())
				}
			}

			if _, ok := entry["time"]; !ok {
				t.Errorf("missing time field in %q", buf.String())
			}
		})
	}
}

func TestNewHandlerLevel(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	logger := slog.New(logging.NewHandler(&buf, slog.LevelInfo))
	logger.DebugContext(t.Context(), "hidden")

	if buf.Len() != 0 {
		t.Errorf("got %q, want nothing below the handler's level", buf.String())
	}
}
