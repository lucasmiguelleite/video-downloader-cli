package youtube

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"youtube-downloader/internal/downloader"

	yt "github.com/kkdai/youtube/v2"
	"github.com/schollz/progressbar/v3"
)

// Client downloads YouTube videos through a YouTubeAPI.
type Client struct {
	client YouTubeAPI
}

// NewClient creates a YouTube downloader backed by api.
func NewClient(api YouTubeAPI) *Client {
	return &Client{client: api}
}

// GetVideo resolves YouTube metadata for url.
func (c *Client) GetVideo(ctx context.Context, url string) (*downloader.Video, error) {
	v, err := c.client.GetVideo(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("get YouTube video %q: %w", url, err)
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
		return fmt.Errorf("get YouTube video %q: %w", video.URL, err)
	}

	formats := v.Formats.WithAudioChannels()
	if len(formats) == 0 {
		return fmt.Errorf("no format with audio found")
	}

	format, err := selectFormat(formats, quality)
	if err != nil {
		return err
	}

	stream, size, err := c.client.GetStream(ctx, v, format)
	if err != nil {
		return fmt.Errorf("get YouTube stream at %s: %w", format.QualityLabel, err)
	}
	defer stream.Close()

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create output file %q: %w", outputPath, err)
	}
	defer f.Close()

	bar := progressbar.DefaultBytes(
		size,
		"Downloading",
	)

	if _, err := io.Copy(f, io.TeeReader(stream, bar)); err != nil {
		return fmt.Errorf("write YouTube stream to %q: %w", outputPath, err)
	}
	return nil
}

func selectFormat(formats yt.FormatList, quality string) (*yt.Format, error) {
	target, err := qualityHeight(quality)
	if err != nil {
		return nil, err
	}
	best := -1
	bestBelow := -1
	selected := -1
	for i := range formats {
		height, err := qualityHeight(formats[i].QualityLabel)
		if err != nil {
			continue
		}
		if height > best {
			best = height
		}
		if height <= target && height > bestBelow {
			bestBelow = height
			selected = i
		}
	}
	if selected < 0 {
		for i := range formats {
			height, err := qualityHeight(formats[i].QualityLabel)
			if err == nil && height == best {
				selected = i
				break
			}
		}
	}
	if selected < 0 {
		return nil, fmt.Errorf("no playable format for requested quality %q", quality)
	}
	return &formats[selected], nil
}

func qualityHeight(quality string) (int, error) {
	height, err := strconv.Atoi(strings.TrimSuffix(quality, "p"))
	if err != nil || height < 1 || !strings.HasSuffix(quality, "p") {
		return 0, fmt.Errorf("invalid quality %q; expected a value such as 720p", quality)
	}
	return height, nil
}
