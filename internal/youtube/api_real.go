package youtube

import (
	"context"
	"io"

	"github.com/kkdai/youtube/v2"
)

// RealAPI adapts the kkdai YouTube client to YouTubeAPI.
type RealAPI struct {
	client youtube.Client
}

// NewRealAPI creates the production YouTube API adapter.
func NewRealAPI() *RealAPI {
	return &RealAPI{}
}

func (r *RealAPI) GetVideo(ctx context.Context, url string) (*youtube.Video, error) {
	return r.client.GetVideoContext(ctx, url)
}

func (r *RealAPI) GetStream(ctx context.Context, v *youtube.Video, f *youtube.Format) (io.ReadCloser, int64, error) {
	return r.client.GetStreamContext(ctx, v, f)
}
