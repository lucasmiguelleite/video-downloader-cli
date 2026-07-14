package twitch

import "context"

type VideoInfo struct {
	Title string
}

type AccessToken struct {
	Value     string
	Signature string
}

type TwitchAPI interface {
	FetchVideoInfo(ctx context.Context, videoID string) (*VideoInfo, error)
	FetchAccessToken(ctx context.Context, videoID string) (*AccessToken, error)
	FetchPlaylist(ctx context.Context, m3u8URL string) (string, error)
	FetchSegment(ctx context.Context, url string) ([]byte, error)
}
