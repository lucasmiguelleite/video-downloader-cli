//go:build !unix

package cli

import (
	"os"
	"testing"
)

func TestIsRunningInBackground_AlwaysFalse(t *testing.T) {
	result, err := IsRunningInBackground(os.Stdin)
	if err != nil {
		t.Fatalf("expected no error: %v", err)
	}

	if result {
		t.Error("expected false on non-unix platforms")
	}
}
