package fs

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DefaultDownloadDir returns the user's conventional Downloads directory.
func DefaultDownloadDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Downloads"), nil
}

// SafeFilename returns a portable filename derived from title with extension.
// Path separators and control characters are replaced so title cannot escape a
// caller-selected directory.
func SafeFilename(title, extension string) string {
	title = strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || r == 0 || r < 32 {
			return '_'
		}
		return r
	}, title)
	title = strings.Trim(strings.TrimSpace(title), ".")
	if title == "" {
		title = "video"
	}
	return title + extension
}

// MkdirAll creates dir (and any missing parents) with mode 0755. It is a
// no-op (returns nil) when dir already exists.
// MkdirAll creates dir and any missing parent directories.
func MkdirAll(dir string) error {
	return os.MkdirAll(dir, 0755)
}

// Create makes filename in dir, creating dir when necessary.
func Create(dir, filename string) (*os.File, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return os.Create(filepath.Join(dir, filename))
}

// RemuxToMP4 replaces path with an ffmpeg-remuxed MP4 stream.
func RemuxToMP4(ctx context.Context, path string) error {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg is required to convert HLS segments to MP4. Install it with: brew install ffmpeg")
	}

	ext := filepath.Ext(path)
	tmpPath := strings.TrimSuffix(path, ext) + "_remuxed" + ext

	cmd := exec.CommandContext(ctx, "ffmpeg", "-y", "-i", path, "-c", "copy", tmpPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("ffmpeg failed: %s: %w", string(output), err)
	}

	if err := os.Remove(path); err != nil {
		return err
	}

	return os.Rename(tmpPath, path)
}
