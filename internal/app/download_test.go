package app

import (
	"os"
	"path/filepath"
	"testing"
	"youtube-downloader/internal/downloader"
)

type mockService struct{}

func (m *mockService) GetVideo(url string) (*downloader.Video, error) {
	return &downloader.Video{Title: "test"}, nil
}

func (m *mockService) Download(v *downloader.Video, q string) ([]byte, error) {
	return []byte("fake"), nil
}

func TestDownloaderUseCase(t *testing.T) {
	dir := t.TempDir()

	service := &mockService{}
	uc := NewDownloadUseCaseWithDir(service, dir)

	path, err := uc.Execute("url", "720p")

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
