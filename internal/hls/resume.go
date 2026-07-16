package hls

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// Manifest records the segment list and how many leading segments have been
// durably written to the output file. It is persisted as JSON in a `.part`
// sidecar next to the output so an interrupted download can be resumed.
type Manifest struct {
	Segments  []string `json:"segments"`
	Completed int      `json:"completed"`
}

// LoadManifest reads a resume manifest from path. It returns (nil, nil) when
// the sidecar does not exist (the common case — fresh download). Any other
// read or parse error is returned.
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// Save writes the manifest to path as indented JSON. The write is atomic: it
// goes to a temp file first, then renames over the target.
func (m *Manifest) Save(path string) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return atomicWrite(path, data, 0o644)
}

// Delete removes the sidecar. It is idempotent — deleting a non-existent file
// is not an error.
func (m *Manifest) Delete(path string) error {
	err := os.Remove(path)
	if err == nil || errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

// Matches reports whether the given segment list equals the manifest's. This
// is the resume-safety check: if the segment list changed (e.g. the live VOD
// was re-encoded, or a different URL was used), we must not reuse the partial
// file.
func (m *Manifest) Matches(segments []string) bool {
	if len(m.Segments) != len(segments) {
		return false
	}
	for i := range segments {
		if m.Segments[i] != segments[i] {
			return false
		}
	}
	return true
}

// atomicWrite writes data to a sibling temp file then renames it over path.
func atomicWrite(path string, data []byte, perm os.FileMode) error {
	dir := dirOf(path)
	tmp, err := os.CreateTemp(dir, ".resume-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op if rename succeeded

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, perm); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

// dirOf returns the directory part of path, falling back to "." so
// CreateTemp always has a valid dir even for a bare filename.
func dirOf(path string) string {
	if d := filepath.Dir(path); d != "" {
		return d
	}
	return "."
}
