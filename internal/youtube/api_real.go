package youtube

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/kkdai/youtube/v2"
)

// RealAPI adapts the kkdai YouTube client to YouTubeAPI.
type RealAPI struct {
	client youtube.Client
}

// NewRealAPI creates the production YouTube API adapter.
func NewRealAPI() *RealAPI {
	return NewRealAPIWithHTTPClient(&http.Client{Timeout: 30 * time.Second})
}

// NewRealAPIWithHTTPClient creates a YouTube API adapter using client.
func NewRealAPIWithHTTPClient(client *http.Client) *RealAPI {
	if client == nil {
		client = http.DefaultClient
	}
	return &RealAPI{client: youtube.Client{HTTPClient: client}}
}

func (r *RealAPI) GetVideo(ctx context.Context, url string) (*youtube.Video, error) {
	return r.client.GetVideoContext(ctx, url)
}

func (r *RealAPI) GetStream(ctx context.Context, v *youtube.Video, f *youtube.Format) (io.ReadCloser, int64, error) {
	return r.client.GetStreamContext(ctx, v, f)
}
