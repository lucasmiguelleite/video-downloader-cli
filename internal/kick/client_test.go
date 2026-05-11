package kick

import (
	"bytes"
	"testing"

	"youtube-downloader/internal/downloader"
)

type mockAPI struct {
	videoInfo *VideoInfo
	playlists map[string]string
	segments  map[string][]byte
	videoErr  error
}

func (m *mockAPI) FetchVideoInfo(videoID string) (*VideoInfo, error) {
	if m.videoErr != nil {
		return nil, m.videoErr
	}
	return m.videoInfo, nil
}

func (m *mockAPI) FetchPlaylist(m3u8URL string) (string, error) {
	if m.playlists != nil {
		if p, ok := m.playlists[m3u8URL]; ok {
			return p, nil
		}
	}
	return "", nil
}

func (m *mockAPI) FetchSegment(url string) ([]byte, error) {
	if data, ok := m.segments[url]; ok {
		return data, nil
	}
	return []byte("segment-data"), nil
}

func TestGetVideo(t *testing.T) {
	api := &mockAPI{
		videoInfo: &VideoInfo{
			Title:     "Test Stream",
			SourceURL: "https://cdn.example.com/video.m3u8",
		},
	}

	client := NewClient(api)
	video, err := client.GetVideo("https://kick.com/nitrao/videos/7b0877f1-983d-4c5f-932e-cda88e8196f1")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if video.Title != "Test Stream" {
		t.Errorf("expected title 'Test Stream', got '%s'", video.Title)
	}

	if video.URL != "https://cdn.example.com/video.m3u8" {
		t.Errorf("expected source URL, got '%s'", video.URL)
	}
}

func TestGetVideo_InvalidURL(t *testing.T) {
	api := &mockAPI{}
	client := NewClient(api)

	_, err := client.GetVideo("https://youtube.com/watch?v=xxx")

	if err == nil {
		t.Error("expected error for non-Kick URL")
	}
}

func TestDownload_MediaPlaylist(t *testing.T) {
	playlist := `#EXTM3U
#EXTINF:10.0,
https://cdn.example.com/seg1.ts
#EXTINF:10.0,
https://cdn.example.com/seg2.ts
#EXT-X-ENDLIST`

	api := &mockAPI{
		playlists: map[string]string{
			"https://cdn.example.com/video.m3u8": playlist,
		},
		segments: map[string][]byte{
			"https://cdn.example.com/seg1.ts": []byte("data1"),
			"https://cdn.example.com/seg2.ts": []byte("data2"),
		},
	}

	client := NewClient(api)
	var buf bytes.Buffer
	err := client.Download(&downloader.Video{
		Title: "Test",
		URL:   "https://cdn.example.com/video.m3u8",
	}, "720p", &buf)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	expected := "data1data2"
	if buf.String() != expected {
		t.Errorf("expected '%s', got '%s'", expected, buf.String())
	}
}

func TestDownload_MasterPlaylist(t *testing.T) {
	master := `#EXTM3U
#EXT-X-STREAM-INF:BANDWIDTH=4000000,RESOLUTION=1920x1080
https://cdn.example.com/1080p/index.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=2000000,RESOLUTION=1280x720
https://cdn.example.com/720p/index.m3u8`

	media := `#EXTM3U
#EXTINF:10.0,
https://cdn.example.com/720p/seg1.ts
#EXTINF:10.0,
https://cdn.example.com/720p/seg2.ts
#EXT-X-ENDLIST`

	api := &mockAPI{
		playlists: map[string]string{
			"https://cdn.example.com/master.m3u8":    master,
			"https://cdn.example.com/720p/index.m3u8": media,
		},
		segments: map[string][]byte{
			"https://cdn.example.com/720p/seg1.ts": []byte("data1"),
			"https://cdn.example.com/720p/seg2.ts": []byte("data2"),
		},
	}

	client := NewClient(api)
	var buf bytes.Buffer
	err := client.Download(&downloader.Video{
		Title: "Test",
		URL:   "https://cdn.example.com/master.m3u8",
	}, "720p", &buf)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	expected := "data1data2"
	if buf.String() != expected {
		t.Errorf("expected '%s', got '%s'", expected, buf.String())
	}
}
