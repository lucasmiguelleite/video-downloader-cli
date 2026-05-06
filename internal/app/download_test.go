package app

import (
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
	service := &mockService{}
	uc := NewDownloadUseCase(service)

	err := uc.Execute("url", "720p")

	if err != nil {
		t.Fatalf("Não esperava erro, mas recebeu: %v", err)
	}
}
