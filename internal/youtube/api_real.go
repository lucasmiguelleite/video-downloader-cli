package youtube

import (
	"io"

	"github.com/kkdai/youtube/v2"
)

type RealAPI struct {
	client youtube.Client
}

func NewRealAPI() *RealAPI {
	return &RealAPI{}
}

func (r *RealAPI) GetVideo(url string) (*youtube.Video, error) {
	return r.client.GetVideo(url)
}

func (r *RealAPI) GetStream(v *youtube.Video, f *youtube.Format) (io.ReadCloser, int64, error) {
	return r.client.GetStream(v, f)
}
