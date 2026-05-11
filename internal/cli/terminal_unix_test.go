//go:build unix

package cli

import (
	"os"
	"testing"
)

func TestIsRunningInBackground_WithNonTerminal(t *testing.T) {
	f, err := os.CreateTemp("", "stdin")
	if err != nil {
		t.Fatalf("error creating temp file: %v", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	result, err := IsRunningInBackground(f)
	if err != nil {
		t.Fatalf("expected no error with file: %v", err)
	}

	if result {
		t.Error("expected false for non-terminal file")
	}
}

func TestIsRunningInBackground_WithPipe(t *testing.T) {
	_, err := IsRunningInBackground(os.Stdin)

	if err != nil {
		t.Logf("stdin is not a tty (expected in go test): %v", err)
	}
}
