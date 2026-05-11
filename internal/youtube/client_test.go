package youtube

import (
	"bytes"
	"io"
	"testing"

	"youtube-downloader/internal/downloader"

	"github.com/kkdai/youtube/v2"
)

type mockAPI struct{}

func (m *mockAPI) GetVideo(url string) (*youtube.Video, error) {
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

func (m *mockAPI) GetStream(v *youtube.Video, f *youtube.Format) (io.ReadCloser, int64, error) {
	data := []byte("fake video data")
	return io.NopCloser(bytes.NewReader(data)), int64(len(data)), nil
}

func TestGetVideo(t *testing.T) {
	client := NewClient(&mockAPI{})

	video, err := client.GetVideo("http://test")

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

	var buf bytes.Buffer
	err := client.Download(video, "720p", &buf)

	if err != nil {
		t.Fatalf("expected no error: %v", err)
	}

	if buf.String() != "fake video data" {
		t.Errorf("invalid content")
	}
}
