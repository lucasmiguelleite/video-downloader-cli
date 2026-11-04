package youtube

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"youtube-downloader/internal/downloader"

	"github.com/kkdai/youtube/v2"
)

type mockAPI struct{}

func (m *mockAPI) GetVideo(ctx context.Context, url string) (*youtube.Video, error) {
	return &youtube.Video{
		Title: "test-video",
		Formats: youtube.FormatList{
			{
				QualityLabel:  "720p",
				MimeType:      "video/mp4",
				AudioChannels: 2,
			},
		},
	}, nil
}

func (m *mockAPI) GetStream(ctx context.Context, v *youtube.Video, f *youtube.Format) (io.ReadCloser, int64, error) {
	data := []byte("fake video data")
	return io.NopCloser(bytes.NewReader(data)), int64(len(data)), nil
}

func TestGetVideo(t *testing.T) {
	client := NewClient(&mockAPI{})

	video, err := client.GetVideo(context.Background(), "http://test")

	if err != nil {
		t.Fatalf("expected no error")
	}

	if video.Title != "test-video" {
		t.Errorf("incorrect title")
	}
}

func TestDownload(t *testing.T) {
	client := NewClient(&mockAPI{})

	video := &downloader.Video{
		URL: "http://test",
	}

	path := filepath.Join(t.TempDir(), "out.mp4")
	err := client.Download(context.Background(), video, "720p", path)

	if err != nil {
		t.Fatalf("expected no error: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("file not written: %v", err)
	}
	if string(content) != "fake video data" {
		t.Errorf("invalid content")
	}
}

func TestSelectFormat_PrefersRequestedQuality(t *testing.T) {
	formats := youtube.FormatList{{QualityLabel: "1080p"}, {QualityLabel: "720p"}, {QualityLabel: "480p"}}
	format, err := selectFormat(formats, "720p")
	if err != nil {
		t.Fatalf("selectFormat: %v", err)
	}
	if format.QualityLabel != "720p" {
		t.Errorf("selected %q, want 720p", format.QualityLabel)
	}
}

func TestSelectFormat_FallsBackToHighestAvailableBelowRequested(t *testing.T) {
	formats := youtube.FormatList{{QualityLabel: "1080p"}, {QualityLabel: "480p"}}
	format, err := selectFormat(formats, "720p")
	if err != nil {
		t.Fatalf("selectFormat: %v", err)
	}
	if format.QualityLabel != "480p" {
		t.Errorf("selected %q, want 480p", format.QualityLabel)
	}
}

func TestSelectFormat_InvalidQuality(t *testing.T) {
	_, err := selectFormat(youtube.FormatList{{QualityLabel: "720p"}}, "highest")
	if err == nil {
		t.Fatal("expected invalid quality error")
	}
}

func TestSelectFormat_ReturnsErrorWhenNoPlayableFormatExists(t *testing.T) {
	_, err := selectFormat(youtube.FormatList{{QualityLabel: "audio"}}, "720p")
	if err == nil {
		t.Fatal("expected no playable format error")
	}
}
