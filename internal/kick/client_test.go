package kick

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"youtube-downloader/internal/downloader"
)

type mockAPI struct {
	videoInfo *VideoInfo
	playlists map[string]string
	segments  map[string][]byte
	videoErr  error
}

func (m *mockAPI) FetchVideoInfo(ctx context.Context, videoID string) (*VideoInfo, error) {
	if m.videoErr != nil {
		return nil, m.videoErr
	}
	return m.videoInfo, nil
}

func (m *mockAPI) FetchPlaylist(ctx context.Context, m3u8URL string) (string, error) {
	if m.playlists != nil {
		if p, ok := m.playlists[m3u8URL]; ok {
			return p, nil
		}
	}
	return "", nil
}

func (m *mockAPI) FetchSegment(ctx context.Context, url string) ([]byte, error) {
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
	video, err := client.GetVideo(context.Background(), "https://kick.com/nitrao/videos/7b0877f1-983d-4c5f-932e-cda88e8196f1")

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

	_, err := client.GetVideo(context.Background(), "https://youtube.com/watch?v=xxx")

	if err == nil {
		t.Error("expected error for non-Kick URL")
	}
}

func TestGetVideo_RejectsLookalikeHost(t *testing.T) {
	_, err := NewClient(&mockAPI{}).GetVideo(context.Background(), "https://notkick.com/a/videos/7b0877f1-983d-4c5f-932e-cda88e8196f1")
	if err == nil {
		t.Fatal("expected lookalike host error")
	}
}

func TestGetVideo_RejectsHTTPURL(t *testing.T) {
	_, err := NewClient(&mockAPI{}).GetVideo(context.Background(), "http://kick.com/a/videos/7b0877f1-983d-4c5f-932e-cda88e8196f1")
	if err == nil {
		t.Fatal("expected HTTP scheme error")
	}
}

func TestGetVideo_RejectsEncodedPath(t *testing.T) {
	_, err := NewClient(&mockAPI{}).GetVideo(context.Background(), "https://kick.com/a/videos/%37b0877f1-983d-4c5f-932e-cda88e8196f1")
	if err == nil {
		t.Fatal("expected encoded path error")
	}
}

func TestGetVideo_NormalizesUppercaseUUID(t *testing.T) {
	id, err := extractVideoID("https://kick.com/a/videos/7B0877F1-983D-4C5F-932E-CDA88E8196F1")
	if err != nil {
		t.Fatal(err)
	}
	if id != "7b0877f1-983d-4c5f-932e-cda88e8196f1" {
		t.Errorf("video ID = %q", id)
	}
}

func TestDownload_MediaPlaylist(t *testing.T) {
	playlist := `#EXTM3U
#EXTINF:10.0,
https://cdn.example.com/seg1.ts
#EXTINF:10.0,
https://cdn.example.com/seg2.ts
#EXTINF:10.0,
https://cdn.example.com/seg3.ts
#EXT-X-ENDLIST`

	api := &mockAPI{
		playlists: map[string]string{
			"https://cdn.example.com/video.m3u8": playlist,
		},
		segments: map[string][]byte{
			"https://cdn.example.com/seg1.ts": []byte("data1"),
			"https://cdn.example.com/seg2.ts": []byte("data2"),
			"https://cdn.example.com/seg3.ts": []byte("data3"),
		},
	}

	client := NewClient(api)
	path := filepath.Join(t.TempDir(), "out.ts")
	err := client.Download(context.Background(), &downloader.Video{
		Title: "Test",
		URL:   "https://cdn.example.com/video.m3u8",
	}, "720p", path)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	got, _ := os.ReadFile(path)
	expected := "data1data2data3"
	if string(got) != expected {
		t.Errorf("expected '%s', got '%s'", expected, string(got))
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
			"https://cdn.example.com/master.m3u8":     master,
			"https://cdn.example.com/720p/index.m3u8": media,
		},
		segments: map[string][]byte{
			"https://cdn.example.com/720p/seg1.ts": []byte("data1"),
			"https://cdn.example.com/720p/seg2.ts": []byte("data2"),
		},
	}

	client := NewClient(api)
	path := filepath.Join(t.TempDir(), "out.ts")
	err := client.Download(context.Background(), &downloader.Video{
		Title: "Test",
		URL:   "https://cdn.example.com/master.m3u8",
	}, "720p", path)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	got, _ := os.ReadFile(path)
	expected := "data1data2"
	if string(got) != expected {
		t.Errorf("expected '%s', got '%s'", expected, string(got))
	}
}
