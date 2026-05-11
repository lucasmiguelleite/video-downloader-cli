package downloader

import "io"

type Video struct {
	Title string
	URL   string
}

type Service interface {
	GetVideo(url string) (*Video, error)
	Download(video *Video, quality string, w io.Writer) error
}
