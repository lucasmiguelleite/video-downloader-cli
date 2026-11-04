package youtube

import (
	"context"
	"io"

	"github.com/kkdai/youtube/v2"
)

// YouTubeAPI abstracts the external YouTube client.
type YouTubeAPI interface {
	GetVideo(ctx context.Context, url string) (*youtube.Video, error)
	GetStream(ctx context.Context, video *youtube.Video, format *youtube.Format) (io.ReadCloser, int64, error)
}
