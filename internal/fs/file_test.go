package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSave(t *testing.T) {
	dir := t.TempDir()

	filename := filepath.Join(dir, "test.mp4")
	data := []byte("fake data")

	err := Save(filename, data)
	if err != nil {
		t.Fatalf("erro ao salvar: %v", err)
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("erro ao ler: %v", err)
	}

	if string(content) != "fake data" {
		t.Errorf("conteúdo incorreto")
	}
}
