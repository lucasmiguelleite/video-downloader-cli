package twitch

import "context"

// VideoInfo contains Twitch metadata needed for a VOD download.
type VideoInfo struct {
	Title string
}

// AccessToken authorizes requests to a Twitch VOD playlist.
type AccessToken struct {
	Value     string
	Signature string
}

// TwitchAPI abstracts Twitch's metadata and HLS endpoints.
type TwitchAPI interface {
	FetchVideoInfo(ctx context.Context, videoID string) (*VideoInfo, error)
	FetchAccessToken(ctx context.Context, videoID string) (*AccessToken, error)
	FetchPlaylist(ctx context.Context, m3u8URL string) (string, error)
	FetchSegment(ctx context.Context, url string) ([]byte, error)
}
