package kick

import (
	"context"
	"io"
)

// VideoInfo contains Kick metadata needed to start an HLS download.
type VideoInfo struct {
	Title     string
	SourceURL string
}

// KickAPI abstracts Kick's metadata and HLS endpoints.
type KickAPI interface {
	FetchVideoInfo(ctx context.Context, videoID string) (*VideoInfo, error)
	FetchPlaylist(ctx context.Context, m3u8URL string) (string, error)
	FetchSegment(ctx context.Context, url string) (io.ReadCloser, error)
}
