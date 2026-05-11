package fs

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func DefaultDownloadDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Downloads"), nil
}

func Create(dir, filename string) (*os.File, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return os.Create(filepath.Join(dir, filename))
}

func RemuxToMP4(path string) error {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg is required to convert HLS segments to MP4. Install it with: brew install ffmpeg")
	}

	ext := filepath.Ext(path)
	tmpPath := strings.TrimSuffix(path, ext) + "_remuxed" + ext

	cmd := exec.Command("ffmpeg", "-y", "-i", path, "-c", "copy", tmpPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("ffmpeg failed: %s: %w", string(output), err)
	}

	if err := os.Remove(path); err != nil {
		return err
	}

	return os.Rename(tmpPath, path)
}
