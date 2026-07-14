package downloader

import "context"

type Video struct {
	Title string
	URL   string
}

type Service interface {
	GetVideo(ctx context.Context, url string) (*Video, error)
	// Download fetches the video and writes it to outputPath. outputPath is a
	// filesystem path (not an io.Writer) so the implementation can keep a
	// resume sidecar next to the output file.
	Download(ctx context.Context, video *Video, quality, outputPath string) error
}
