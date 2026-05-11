package youtube

import (
	"fmt"
	"io"
	"youtube-downloader/internal/downloader"

	"github.com/schollz/progressbar/v3"
)

type Client struct {
	client YouTubeAPI
}

func NewClient(api YouTubeAPI) *Client {
	return &Client{client: api}
}

func (c *Client) GetVideo(url string) (*downloader.Video, error) {
	v, err := c.client.GetVideo(url)
	if err != nil {
		return nil, err
	}

	return &downloader.Video{
		Title: v.Title,
		URL:   url,
	}, nil
}

func (c *Client) Download(video *downloader.Video, quality string, w io.Writer) error {
	v, err := c.client.GetVideo(video.URL)
	if err != nil {
		return err
	}

	formats := v.Formats.WithAudioChannels()
	if len(formats) == 0 {
		return fmt.Errorf("no format with audio found")
	}

	format := &formats[0]

	stream, size, err := c.client.GetStream(v, format)
	if err != nil {
		return err
	}
	defer stream.Close()

	bar := progressbar.DefaultBytes(
		size,
		"Downloading",
	)

	_, err = io.Copy(w, io.TeeReader(stream, bar))
	if err != nil {
		return err
	}

	fmt.Println("\nDownload complete!")
	return nil
}
