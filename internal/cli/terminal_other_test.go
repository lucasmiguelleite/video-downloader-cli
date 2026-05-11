//go:build !unix

package cli

import (
	"os"
	"testing"
)

func TestIsRunningInBackground_AlwaysFalse(t *testing.T) {
	result, err := IsRunningInBackground(os.Stdin)
	if err != nil {
		t.Fatalf("não esperava erro: %v", err)
	}

	if result {
		t.Error("esperava false em plataformas não-unix")
	}
}
