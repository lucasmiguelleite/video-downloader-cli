package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreate(t *testing.T) {
	dir := t.TempDir()

	f, err := Create(dir, "test.mp4")
	if err != nil {
		t.Fatalf("error creating file: %v", err)
	}

	_, err = f.Write([]byte("fake data"))
	if err != nil {
		t.Fatalf("error writing: %v", err)
	}
	f.Close()

	content, err := os.ReadFile(filepath.Join(dir, "test.mp4"))
	if err != nil {
		t.Fatalf("error reading: %v", err)
	}

	if string(content) != "fake data" {
		t.Errorf("incorrect content")
	}
}

func TestCreate_CreatesDir(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "subdir", "nested")

	f, err := Create(dir, "test.mp4")
	if err != nil {
		t.Fatalf("error creating file in nonexistent directory: %v", err)
	}
	f.Close()

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
