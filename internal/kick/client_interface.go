package kick

import "context"

type VideoInfo struct {
	Title     string
	SourceURL string
}

type KickAPI interface {
	FetchVideoInfo(ctx context.Context, videoID string) (*VideoInfo, error)
	FetchPlaylist(ctx context.Context, m3u8URL string) (string, error)
	FetchSegment(ctx context.Context, url string) ([]byte, error)
}
