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

func (c *Client) Download(video *downloader.Video, quality string) ([]byte, error) {
	v, err := c.client.GetVideo(video.URL)
	if err != nil {
		return nil, err
	}

	// get formats with audio
	formats := v.Formats.WithAudioChannels()
	if len(formats) == 0 {
		return nil, fmt.Errorf("no format with audio found")
	}

	format := &formats[0]

	// 👇 aqui muda: agora pegamos o size também
	stream, size, err := c.client.GetStream(v, format)
	if err != nil {
		return nil, err
	}

	// 👇 cria barra de progresso
	bar := progressbar.DefaultBytes(
		size,
		"Downloading",
	)

	// 👇 junta stream + barra
	reader := io.TeeReader(stream, bar)

	// 👇 continua usando memória (por enquanto)
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	fmt.Println("\nDownload complete!")

	return data, nil
}
