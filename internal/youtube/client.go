package youtube

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
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

// Download fetches the video and atomically replaces outputPath on success.
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

	bar := progressbar.DefaultBytes(size, "Downloading")
	return writeStreamAtomically(outputPath, io.TeeReader(stream, bar))
}

func writeStreamAtomically(outputPath string, stream io.Reader) error {
	temp, err := os.CreateTemp(filepath.Dir(outputPath), ".youtube-*")
	if err != nil {
		return fmt.Errorf("create temporary output for %q: %w", outputPath, err)
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if _, err := io.Copy(temp, stream); err != nil {
		_ = temp.Close()
		return fmt.Errorf("write YouTube stream to %q: %w", outputPath, err)
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return fmt.Errorf("sync YouTube output %q: %w", outputPath, err)
	}
	if err := temp.Chmod(0o644); err != nil {
		_ = temp.Close()
		return fmt.Errorf("set YouTube output permissions %q: %w", outputPath, err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close YouTube output %q: %w", outputPath, err)
	}
	if err := os.Rename(tempPath, outputPath); err != nil {
		return fmt.Errorf("replace YouTube output %q: %w", outputPath, err)
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
