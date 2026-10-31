// Package logging writes structured logs in the JSON format that Google Cloud
// Logging understands, so that each entry's severity and message are shown as
// such rather than as part of a line of text.
package logging

import (
	"io"
	"log/slog"
)

// NewHandler returns a handler that writes each record to w as a line of JSON,
// with the level in a "severity" field (using Cloud Logging's severity names)
// and the message in a "message" field. Other attributes are written as usual.
func NewHandler(w io.Writer, level slog.Leveler) slog.Handler {
	return slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level:       level,
		ReplaceAttr: replaceAttr,
	})
}

func replaceAttr(groups []string, attr slog.Attr) slog.Attr {
	// The built-in keys are only ever at the top level, so an attribute in a
	// group with the same key is left alone.
	if len(groups) > 0 {
		return attr
	}

	switch attr.Key {
	case slog.LevelKey:
		level, ok := attr.Value.Any().(slog.Level)
		if !ok {
			return attr
		}

		return slog.String("severity", severity(level))
	case slog.MessageKey:
		return slog.Attr{Key: "message", Value: attr.Value}
	default:
		return attr
	}
}

// severity maps a level to the Cloud Logging severity it falls within. Levels
// between the standard ones (e.g. slog.LevelWarn+2) get the severity of the
// standard level below them.
func severity(level slog.Level) string {
	switch {
	case level < slog.LevelInfo:
		return "DEBUG"
	case level < slog.LevelWarn:
		return "INFO"
	case level < slog.LevelError:
		return "WARNING"
	default:
		return "ERROR"
	}
}
