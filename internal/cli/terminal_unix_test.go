//go:build unix

package cli

import (
	"os"
	"testing"
)

func TestIsRunningInBackground_WithNonTerminal(t *testing.T) {
	// Arquivo normal não é um tty → IoctlGetInt retorna ENOTTY
	// A função trata esse caso e retorna (false, nil)
	f, err := os.CreateTemp("", "stdin")
	if err != nil {
		t.Fatalf("erro ao criar arquivo temporário: %v", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	result, err := IsRunningInBackground(f)
	if err != nil {
		t.Fatalf("não esperava erro com arquivo: %v", err)
	}

	if result {
		t.Error("esperava false para arquivo não-terminal")
	}
}

func TestIsRunningInBackground_WithPipe(t *testing.T) {
	// No go test, stdin é um pipe — não é um tty.
	// A função pode retornar erro dependendo do ambiente,
	// o importante é que não panique.
	_, err := IsRunningInBackground(os.Stdin)

	if err != nil {
		t.Logf("stdin não é tty (esperado em go test): %v", err)
	}
}
