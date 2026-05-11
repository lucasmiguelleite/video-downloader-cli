package kick

type VideoInfo struct {
	Title     string
	SourceURL string
}

type KickAPI interface {
	FetchVideoInfo(videoID string) (*VideoInfo, error)
	FetchPlaylist(m3u8URL string) (string, error)
	FetchSegment(url string) ([]byte, error)
}
