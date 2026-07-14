package youtube

import (
	"context"
	"fmt"
	"io"
	"os"
	"youtube-downloader/internal/downloader"

	"github.com/schollz/progressbar/v3"
)

type Client struct {
	client YouTubeAPI
}

func NewClient(api YouTubeAPI) *Client {
	return &Client{client: api}
}

func (c *Client) GetVideo(ctx context.Context, url string) (*downloader.Video, error) {
	v, err := c.client.GetVideo(ctx, url)
	if err != nil {
		return nil, err
	}

	return &downloader.Video{
		Title: v.Title,
		URL:   url,
	}, nil
}

// Download fetches the video and writes it to outputPath. YouTube streams are
// single-file (not segmented), so resumability is not supported here — the
// output is truncated at the start of every run.
func (c *Client) Download(ctx context.Context, video *downloader.Video, quality, outputPath string) error {
	v, err := c.client.GetVideo(ctx, video.URL)
	if err != nil {
		return err
	}

	formats := v.Formats.WithAudioChannels()
	if len(formats) == 0 {
		return fmt.Errorf("no format with audio found")
	}

	format := &formats[0]

	stream, size, err := c.client.GetStream(ctx, v, format)
	if err != nil {
		return err
	}
	defer stream.Close()

	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	bar := progressbar.DefaultBytes(
		size,
		"Downloading",
	)

	_, err = io.Copy(f, io.TeeReader(stream, bar))
	return err
}
