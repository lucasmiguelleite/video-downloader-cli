package twitch

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"youtube-downloader/internal/downloader"
)

type mockAPI struct {
	videoInfo   *VideoInfo
	accessToken *AccessToken
	playlists   map[string]string
	segments    map[string][]byte
	videoErr    error
}

func (m *mockAPI) FetchVideoInfo(ctx context.Context, videoID string) (*VideoInfo, error) {
	if m.videoErr != nil {
		return nil, m.videoErr
	}
	return m.videoInfo, nil
}

func (m *mockAPI) FetchAccessToken(ctx context.Context, videoID string) (*AccessToken, error) {
	return m.accessToken, nil
}

func (m *mockAPI) FetchPlaylist(ctx context.Context, m3u8URL string) (string, error) {
	if m.playlists != nil {
		if p, ok := m.playlists[m3u8URL]; ok {
			return p, nil
		}
	}
	return "", nil
}

func (m *mockAPI) FetchSegment(ctx context.Context, url string) (io.ReadCloser, error) {
	if data, ok := m.segments[url]; ok {
		return io.NopCloser(bytes.NewReader(data)), nil
	}
	return io.NopCloser(strings.NewReader("segment-data")), nil
}

func TestGetVideo(t *testing.T) {
	api := &mockAPI{
		videoInfo: &VideoInfo{Title: "Test Stream"},
	}

	client := NewClient(api)
	video, err := client.GetVideo(context.Background(), "https://www.twitch.tv/videos/1234567890")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if video.Title != "Test Stream" {
		t.Errorf("expected title 'Test Stream', got '%s'", video.Title)
	}
}

func TestGetVideo_PlayerURL(t *testing.T) {
	api := &mockAPI{
		videoInfo: &VideoInfo{Title: "Test"},
	}

	client := NewClient(api)
	video, err := client.GetVideo(context.Background(), "https://player.twitch.tv/?video=1234567890")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if video.Title != "Test" {
		t.Errorf("expected title 'Test', got '%s'", video.Title)
	}
}

func TestGetVideo_InvalidURL(t *testing.T) {
	api := &mockAPI{}
	client := NewClient(api)

	_, err := client.GetVideo(context.Background(), "https://youtube.com/watch?v=xxx")

	if err == nil {
		t.Error("expected error for non-Twitch URL")
	}
}

func TestGetVideo_RejectsLookalikeHost(t *testing.T) {
	_, err := NewClient(&mockAPI{}).GetVideo(context.Background(), "https://eviltwitch.tv/videos/123")
	if err == nil {
		t.Fatal("expected lookalike host error")
	}
}

func TestGetVideo_RejectsHTTPURL(t *testing.T) {
	_, err := NewClient(&mockAPI{}).GetVideo(context.Background(), "http://twitch.tv/videos/123")
	if err == nil {
		t.Fatal("expected HTTP scheme error")
	}
}

func TestGetVideo_RejectsEncodedPath(t *testing.T) {
	_, err := NewClient(&mockAPI{}).GetVideo(context.Background(), "https://twitch.tv/videos/%31%32%33")
	if err == nil {
		t.Fatal("expected encoded path error")
	}
}

func TestGetVideo_ChannelVideoURL(t *testing.T) {
	api := &mockAPI{videoInfo: &VideoInfo{Title: "Test"}}
	video, err := NewClient(api).GetVideo(context.Background(), "https://twitch.tv/channel/video/123")
	if err != nil {
		t.Fatal(err)
	}
	if video.URL != "123" {
		t.Errorf("video ID = %q", video.URL)
	}
}

func TestDownload(t *testing.T) {
	token := &AccessToken{Value: "fake-token", Signature: "fake-sig"}
	usherURL := BuildUsherURL("1234567890", token)

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
		accessToken: token,
		playlists: map[string]string{
			usherURL: master,
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
		URL:   "1234567890",
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

func TestBuildUsherURL(t *testing.T) {
	token := &AccessToken{Value: "mytoken", Signature: "mysig"}
	url := BuildUsherURL("12345", token)

	if url == "" {
		t.Error("expected non-empty URL")
	}

	if len(url) < 30 {
		t.Errorf("URL seems too short: %s", url)
	}
}
