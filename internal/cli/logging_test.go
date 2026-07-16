package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewLogger_DefaultLevelIsInfo(t *testing.T) {
	var buf bytes.Buffer
	log := NewLogger(&buf, false)
	log.Debug("should be hidden")
	log.Info("should be visible")

	out := buf.String()
	if strings.Contains(out, "should be hidden") {
		t.Errorf("debug message should be hidden at Info level, got: %s", out)
	}
	if !strings.Contains(out, "should be visible") {
		t.Errorf("info message should be visible, got: %s", out)
	}
}

func TestNewLogger_VerboseEnablesDebug(t *testing.T) {
	var buf bytes.Buffer
	log := NewLogger(&buf, true)
	log.Debug("now visible")

	out := buf.String()
	if !strings.Contains(out, "now visible") {
		t.Errorf("debug message should be visible in verbose mode, got: %s", out)
	}
}

func TestNewLogger_NilWriterFallsBackToDiscard(t *testing.T) {
	// Should not panic when w is nil (defensive).
	log := NewLogger(nil, false)
	if log == nil {
		t.Fatal("expected non-nil logger")
	}
	log.Info("no panic")
}

func TestNewLogger_OutputIsStructured(t *testing.T) {
	// Verify the text handler produces key=value pairs (not free-form text).
	var buf bytes.Buffer
	log := NewLogger(&buf, true)
	log.Info("downloaded", "url", "https://example.com", "bytes", 42)

	out := buf.String()
	for _, want := range []string{"downloaded", "url=https://example.com", "bytes=42"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output, got: %s", want, out)
		}
	}
}
