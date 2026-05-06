// Package downloader defines core domain models and interfaces.
// This package contains no business logic and is tested indirectly.
package downloader

type Video struct {
	Title string
	URL   string
}

type Service interface {
	GetVideo(url string) (*Video, error)
	Download(video *Video, quality string) ([]byte, error)
}
