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
		t.Fatalf("error saving: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, "test.mp4"))
	if err != nil {
		t.Fatalf("error reading: %v", err)
	}

	if string(content) != "fake data" {
		t.Errorf("incorrect content")
	}
}

func TestSave_CreatesDir(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "subdir", "nested")

	err := Save(dir, "test.mp4", []byte("data"))
	if err != nil {
		t.Fatalf("error saving to nonexistent directory: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "test.mp4")); os.IsNotExist(err) {
		t.Error("file was not created")
	}
}

func TestDefaultDownloadDir(t *testing.T) {
	dir, err := DefaultDownloadDir()
	if err != nil {
		t.Fatalf("error getting download directory: %v", err)
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, "Downloads")

	if dir != expected {
		t.Errorf("expected '%s', got '%s'", expected, dir)
	}
}
