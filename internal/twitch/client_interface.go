package twitch

type VideoInfo struct {
	Title string
}

type AccessToken struct {
	Value     string
	Signature string
}

type TwitchAPI interface {
	FetchVideoInfo(videoID string) (*VideoInfo, error)
	FetchAccessToken(videoID string) (*AccessToken, error)
	FetchPlaylist(m3u8URL string) (string, error)
	FetchSegment(url string) ([]byte, error)
}
