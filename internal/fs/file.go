package fs

import (
	"context"
	"errors"
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

// AvailablePath returns path when unused, or when preserveResume is true and
// an HLS resume sidecar exists. Otherwise it returns a numbered sibling.
func AvailablePath(path string, preserveResume bool) (string, error) {
	if preserveResume {
		if _, err := os.Stat(path + ".part"); err == nil {
			return path, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("stat resume sidecar %q: %w", path+".part", err)
		}
	}
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return path, nil
	} else if err != nil {
		return "", fmt.Errorf("stat output path %q: %w", path, err)
	}
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)
	for i := 1; ; i++ {
		candidate := fmt.Sprintf("%s (%d)%s", base, i, ext)
		if _, err := os.Stat(candidate); errors.Is(err, os.ErrNotExist) {
			return candidate, nil
		} else if err != nil {
			return "", fmt.Errorf("stat output candidate %q: %w", candidate, err)
		}
	}
}

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
		_ = os.Remove(tmpPath)
		return fmt.Errorf("ffmpeg failed: %s: %w", string(output), err)
	}
	if err := replaceFile(path, tmpPath); err != nil {
		return fmt.Errorf("replace original %q with remuxed output: %w", path, err)
	}
	return nil
}

func replaceFile(originalPath, replacementPath string) error {
	if err := os.Rename(replacementPath, originalPath); err == nil {
		return nil
	}
	backup, err := unusedTempPath(filepath.Dir(originalPath), ".remux-backup-*")
	if err != nil {
		return err
	}
	if err := os.Rename(originalPath, backup); err != nil {
		return fmt.Errorf("back up original to %q: %w", backup, err)
	}
	if err := os.Rename(replacementPath, originalPath); err != nil {
		if restoreErr := os.Rename(backup, originalPath); restoreErr != nil {
			return fmt.Errorf("install replacement: %w; restore original from %q: %v", err, backup, restoreErr)
		}
		return fmt.Errorf("install replacement: %w", err)
	}
	if err := os.Remove(backup); err != nil {
		return fmt.Errorf("remove remux backup %q: %w", backup, err)
	}
	return nil
}

func unusedTempPath(dir, pattern string) (string, error) {
	temp, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", err
	}
	path := temp.Name()
	if err := temp.Close(); err != nil {
		_ = os.Remove(path)
		return "", err
	}
	if err := os.Remove(path); err != nil {
		return "", err
	}
	return path, nil
}
