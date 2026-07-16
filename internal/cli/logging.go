package cli

import (
	"io"
	"log/slog"
	"os"
)

// NewLogger builds a structured (slog.TextHandler) logger writing to w.
// When verbose is true, Debug-level messages are emitted; otherwise the
// floor is Info. A nil w falls back to io.Discard (never panics).
//
// The handler writes text key=value records, which keeps stdout clean for
// piping while still being greppable on stderr in normal terminal use.
func NewLogger(w io.Writer, verbose bool) *slog.Logger {
	if w == nil {
		w = io.Discard
	}
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	return slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: level}))
}

// DefaultLogger is a Info-level logger writing to stderr, suitable for code
// paths that don't have a configured logger handed to them.
func DefaultLogger() *slog.Logger {
	return NewLogger(os.Stderr, false)
}
