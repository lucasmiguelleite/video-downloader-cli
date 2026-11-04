package hls

import (
	"os"
	"path/filepath"
	"testing"
)

func TestManifest_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	sidecar := filepath.Join(dir, "video.mp4.part")

	m := &Manifest{
		Segments:  []string{"https://cdn/seg1.ts", "https://cdn/seg2.ts"},
		Completed: 1,
	}
	if err := m.Save(sidecar); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := LoadManifest(sidecar)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	if loaded == nil {
		t.Fatal("expected non-nil manifest")
	}
	if loaded.Completed != 1 {
		t.Errorf("expected Completed=1, got %d", loaded.Completed)
	}
	if loaded.Bytes != 0 {
		t.Errorf("expected Bytes=0, got %d", loaded.Bytes)
	}
	if len(loaded.Segments) != 2 {
		t.Errorf("expected 2 segments, got %d", len(loaded.Segments))
	}
	if loaded.Segments[0] != "https://cdn/seg1.ts" {
		t.Errorf("unexpected segment[0]: %s", loaded.Segments[0])
	}
}

func TestLoadManifest_MissingReturnsNil(t *testing.T) {
	dir := t.TempDir()
	loaded, err := LoadManifest(filepath.Join(dir, "nonexistent.part"))
	if err != nil {
		t.Fatalf("expected nil error for missing file, got: %v", err)
	}
	if loaded != nil {
		t.Errorf("expected nil manifest when file is absent, got %+v", loaded)
	}
}

func TestManifest_DeleteRemovesFile(t *testing.T) {
	dir := t.TempDir()
	sidecar := filepath.Join(dir, "video.mp4.part")

	m := &Manifest{Segments: []string{"x"}, Completed: 0}
	if err := m.Save(sidecar); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if err := m.Delete(sidecar); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := os.Stat(sidecar); !os.IsNotExist(err) {
		t.Errorf("expected file to be deleted, got stat err: %v", err)
	}

	// Deleting again should not error (idempotent).
	if err := m.Delete(sidecar); err != nil {
		t.Errorf("idempotent delete should not error, got: %v", err)
	}
}

func TestManifest_Matches_DetectsSameSegments(t *testing.T) {
	segs := []string{"a", "b", "c"}
	m := &Manifest{Segments: segs, Completed: 1}

	if !m.Matches(segs) {
		t.Error("expected Matches=true for identical segment list")
	}
	if m.Matches([]string{"a", "b", "X"}) {
		t.Error("expected Matches=false for different segment list")
	}
	if m.Matches([]string{"a", "b"}) {
		t.Error("expected Matches=false for shorter list")
	}
}

func TestManifest_Matches_EmptyBoth(t *testing.T) {
	m := &Manifest{Segments: nil}
	if !m.Matches(nil) {
		t.Error("two empty segment lists should match")
	}
}
