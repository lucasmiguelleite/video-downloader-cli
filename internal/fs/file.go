package fs

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const maxFilenameBytes = 255

// ErrPathReserved indicates another process is using a resumable output path.
var ErrPathReserved = errors.New("output path is reserved")

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
		if r == '/' || r == '\\' || r == 0 || r < 32 || strings.ContainsRune(`<>:"|?*`, r) {
			return '_'
		}
		return r
	}, title)
	title = strings.Trim(strings.TrimSpace(title), ".")
	if title == "" {
		title = "video"
	}
	if isReservedFilename(title) {
		title = "_" + title
	}
	title = truncateUTF8(title, maxFilenameBytes-len(extension))
	return title + extension
}

// PathReservation holds an atomically reserved output path.
type PathReservation struct {
	Path     string
	lockPath string
}

// ReservePath atomically reserves path or the first unused numbered sibling.
// A resumable path with a .part sidecar keeps its original name.
func ReservePath(path string, preserveResume bool) (*PathReservation, error) {
	if preserveResume {
		exists, err := exists(path + ".part")
		if err != nil {
			return nil, fmt.Errorf("stat resume sidecar %q: %w", path+".part", err)
		}
		if exists {
			reservation, acquired, err := tryReserve(path)
			if err != nil {
				return nil, err
			}
			if !acquired {
				return nil, fmt.Errorf("%w: %s", ErrPathReserved, path)
			}
			return reservation, nil
		}
	}
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)
	for i := 0; ; i++ {
		candidate := path
		if i > 0 {
			candidate = fmt.Sprintf("%s (%d)%s", base, i, ext)
		}
		reservation, acquired, err := tryReserve(candidate)
		if err != nil {
			return nil, err
		}
		if !acquired {
			continue
		}
		unavailable, err := hasOutputState(candidate)
		if err != nil {
			_ = reservation.Release()
			return nil, err
		}
		if unavailable {
			_ = reservation.Release()
			continue
		}
		return reservation, nil
	}
}

// Release releases the output-path reservation. It is idempotent.
func (r *PathReservation) Release() error {
	if r == nil || r.lockPath == "" {
		return nil
	}
	err := os.Remove(r.lockPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func tryReserve(path string) (*PathReservation, bool, error) {
	lockPath := path + ".lock"
	lock, err := os.OpenFile(lockPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, os.ErrExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("reserve output path %q: %w", path, err)
	}
	if _, err := fmt.Fprintf(lock, "pid=%d\n", os.Getpid()); err != nil {
		_ = lock.Close()
		_ = os.Remove(lockPath)
		return nil, false, fmt.Errorf("write output reservation %q: %w", lockPath, err)
	}
	if err := lock.Close(); err != nil {
		_ = os.Remove(lockPath)
		return nil, false, fmt.Errorf("close output reservation %q: %w", lockPath, err)
	}
	return &PathReservation{Path: path, lockPath: lockPath}, true, nil
}

func hasOutputState(path string) (bool, error) {
	outputExists, err := exists(path)
	if err != nil {
		return false, fmt.Errorf("stat output path %q: %w", path, err)
	}
	partExists, err := exists(path + ".part")
	if err != nil {
		return false, fmt.Errorf("stat resume sidecar %q: %w", path+".part", err)
	}
	return outputExists || partExists, nil
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func truncateUTF8(value string, maxBytes int) string {
	if maxBytes < 1 {
		return ""
	}
	for len(value) > maxBytes {
		_, size := utf8.DecodeLastRuneInString(value)
		value = value[:len(value)-size]
	}
	return value
}

func isReservedFilename(name string) bool {
	upper := strings.ToUpper(name)
	if upper == "CON" || upper == "PRN" || upper == "AUX" || upper == "NUL" {
		return true
	}
	return len(upper) == 4 && (strings.HasPrefix(upper, "COM") || strings.HasPrefix(upper, "LPT")) && upper[3] >= '1' && upper[3] <= '9'
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
