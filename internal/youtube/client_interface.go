package youtube

import (
	"io"

	"github.com/kkdai/youtube/v2"
)

type YouTubeAPI interface {
	GetVideo(url string) (*youtube.Video, error)
	GetStream(video *youtube.Video, format *youtube.Format) (io.ReadCloser, int64, error)
}
