package fs

import (
	"context"
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

func TestSafeFilename_RemovesPathSeparators(t *testing.T) {
	got := SafeFilename("../folder\\video", ".mp4")
	if got != "_folder_video.mp4" {
		t.Errorf("SafeFilename() = %q", got)
	}
}

func TestAvailablePath_AppendsCounter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "video.mp4")
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := AvailablePath(path, false)
	if err != nil {
		t.Fatal(err)
	}
	if got != filepath.Join(dir, "video (1).mp4") {
		t.Errorf("AvailablePath() = %q", got)
	}
}

func TestAvailablePath_PreservesResumePath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "video.mp4")
	if err := os.WriteFile(path, []byte("partial"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path+".part", []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := AvailablePath(path, true)
	if err != nil {
		t.Fatal(err)
	}
	if got != path {
		t.Errorf("AvailablePath() = %q, want resume path %q", got, path)
	}
}

func TestAvailablePath_IgnoresResumeSidecarWhenResumeDisabled(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "video.mp4")
	if err := os.WriteFile(path, []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path+".part", []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := AvailablePath(path, false)
	if err != nil {
		t.Fatal(err)
	}
	if got != filepath.Join(dir, "video (1).mp4") {
		t.Errorf("AvailablePath() = %q", got)
	}
}

func TestRemuxToMP4_NoFFmpeg(t *testing.T) {
	dir := t.TempDir()
	fakePath := filepath.Join(dir, "video.mp4")
	os.WriteFile(fakePath, []byte("fake"), 0644)

	originalPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", originalPath)

	err := RemuxToMP4(context.Background(), fakePath)
	if err == nil {
		t.Error("expected error when ffmpeg is not available")
	}
}
