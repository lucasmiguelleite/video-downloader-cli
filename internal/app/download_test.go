package app

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"youtube-downloader/internal/downloader"
)

type mockService struct {
	getVideoFn func(ctx context.Context, url string) (*downloader.Video, error)
	downloadFn func(ctx context.Context, v *downloader.Video, q, outputPath string) error
}

func (m *mockService) GetVideo(ctx context.Context, url string) (*downloader.Video, error) {
	if m.getVideoFn != nil {
		return m.getVideoFn(ctx, url)
	}
	return &downloader.Video{Title: "test"}, nil
}

func (m *mockService) Download(ctx context.Context, v *downloader.Video, q, outputPath string) error {
	if m.downloadFn != nil {
		return m.downloadFn(ctx, v, q, outputPath)
	}
	return os.WriteFile(outputPath, []byte("fake"), 0o644)
}

func TestDownloaderUseCase(t *testing.T) {
	dir := t.TempDir()

	service := &mockService{}
	uc := NewDownloadUseCaseWithOptions(service, DownloadOptions{Dir: dir})

	path, err := uc.Execute(context.Background(), "url", "720p")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	expectedPath := filepath.Join(dir, "test.mp4")
	if path != expectedPath {
		t.Errorf("expected path '%s', got '%s'", expectedPath, path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("file was not created: %v", err)
	}

	if string(content) != "fake" {
		t.Errorf("incorrect content")
	}
}

// TestDownloaderUseCase_CancellationKeepsPartialFile verifies that when the
// download is cancelled (Ctrl+C), the partial file is kept on disk and the
// use case returns its path rather than an error.
func TestDownloaderUseCase_CancellationKeepsPartialFile(t *testing.T) {
	dir := t.TempDir()

	service := &mockService{
		downloadFn: func(ctx context.Context, v *downloader.Video, q, outputPath string) error {
			// Simulate: write some bytes, then get cancelled.
			if err := os.WriteFile(outputPath, []byte("partial"), 0o644); err != nil {
				return err
			}
			return context.Canceled
		},
	}

	var logBuf bytes.Buffer
	uc := NewDownloadUseCaseWithOptions(service, DownloadOptions{
		Dir:    dir,
		Logger: slog.New(slog.NewTextHandler(&logBuf, nil)),
	})

	path, err := uc.Execute(context.Background(), "url", "720p")

	if err != nil {
		t.Fatalf("expected no error on cancellation, got: %v", err)
	}
	expectedPath := filepath.Join(dir, "test.mp4")
	if path != expectedPath {
		t.Errorf("expected partial path '%s', got '%s'", expectedPath, path)
	}

	// Partial content must still be on disk.
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("partial file should be kept: %v", err)
	}
	if string(content) != "partial" {
		t.Errorf("expected partial content 'partial', got %q", string(content))
	}

	// And a log message about the interruption should mention the path.
	if !bytes.Contains(logBuf.Bytes(), []byte("interrupted")) {
		t.Errorf("expected 'interrupted' in logs, got: %s", logBuf.String())
	}
}

// TestDownloaderUseCase_NonCancellationErrorReturnsError verifies that genuine
// errors (not cancellation) still surface as errors.
func TestDownloaderUseCase_NonCancellationErrorReturnsError(t *testing.T) {
	dir := t.TempDir()
	errSentinel := errors.New("network down")

	service := &mockService{
		downloadFn: func(ctx context.Context, v *downloader.Video, q, outputPath string) error {
			return errSentinel
		},
	}

	uc := NewDownloadUseCaseWithOptions(service, DownloadOptions{Dir: dir})

	_, err := uc.Execute(context.Background(), "url", "720p")
	if !errors.Is(err, errSentinel) {
		t.Errorf("expected sentinel error, got: %v", err)
	}
}

func TestDownloaderUseCase_SanitizesVideoTitle(t *testing.T) {
	dir := t.TempDir()
	service := &mockService{getVideoFn: func(context.Context, string) (*downloader.Video, error) {
		return &downloader.Video{Title: "../../outside"}, nil
	}}

	path, err := NewDownloadUseCaseWithOptions(service, DownloadOptions{Dir: dir}).Execute(context.Background(), "url", "720p")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if filepath.Dir(path) != dir {
		t.Fatalf("download escaped output directory: %q", path)
	}
	if filepath.Base(path) != "_.._outside.mp4" {
		t.Errorf("unexpected sanitized name: %q", filepath.Base(path))
	}
}

func TestDownloaderUseCase_DoesNotOverwriteExistingDownload(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "test.mp4")
	if err := os.WriteFile(existing, []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}
	path, err := NewDownloadUseCaseWithOptions(&mockService{}, DownloadOptions{Dir: dir}).Execute(context.Background(), "url", "720p")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if path != filepath.Join(dir, "test (1).mp4") {
		t.Errorf("path = %q", path)
	}
	content, err := os.ReadFile(existing)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "existing" {
		t.Errorf("existing content was overwritten: %q", content)
	}
}
