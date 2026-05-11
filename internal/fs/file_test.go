package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSave(t *testing.T) {
	dir := t.TempDir()

	err := Save(dir, "test.mp4", []byte("fake data"))
	if err != nil {
		t.Fatalf("erro ao salvar: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, "test.mp4"))
	if err != nil {
		t.Fatalf("erro ao ler: %v", err)
	}

	if string(content) != "fake data" {
		t.Errorf("conteúdo incorreto")
	}
}

func TestSave_CreatesDir(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "subdir", "nested")

	err := Save(dir, "test.mp4", []byte("data"))
	if err != nil {
		t.Fatalf("erro ao salvar em diretório inexistente: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "test.mp4")); os.IsNotExist(err) {
		t.Error("arquivo não foi criado")
	}
}

func TestDefaultDownloadDir(t *testing.T) {
	dir, err := DefaultDownloadDir()
	if err != nil {
		t.Fatalf("erro ao obter diretório de downloads: %v", err)
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, "Downloads")

	if dir != expected {
		t.Errorf("esperava '%s', recebeu '%s'", expected, dir)
	}
}
