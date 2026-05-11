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
		t.Fatalf("Não esperava erro, mas recebeu: %v", err)
	}

	expectedPath := filepath.Join(dir, "test.mp4")
	if path != expectedPath {
		t.Errorf("caminho esperado '%s', recebeu '%s'", expectedPath, path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("arquivo não foi criado: %v", err)
	}

	if string(content) != "fake" {
		t.Errorf("conteúdo incorreto")
	}
}
