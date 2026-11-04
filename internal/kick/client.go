package kick

import (
	"context"
	"fmt"
	"regexp"

	"youtube-downloader/internal/downloader"
	"youtube-downloader/internal/hls"
)

var videoIDRegex = regexp.MustCompile(`kick\.com/\w+/videos/([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})`)

// Client downloads Kick VODs through a KickAPI.
type Client struct {
	api         KickAPI
	concurrency int
}

// NewClient creates a Kick downloader with the default concurrency.
func NewClient(api KickAPI) *Client {
	return &Client{api: api, concurrency: 20}
}

// NewClientWithConcurrency creates a Kick downloader with bounded concurrency.
func NewClientWithConcurrency(api KickAPI, concurrency int) *Client {
	if concurrency < 1 {
		concurrency = 1
	}
	return &Client{api: api, concurrency: concurrency}
}

// GetVideo resolves Kick VOD metadata from url.
func (c *Client) GetVideo(ctx context.Context, url string) (*downloader.Video, error) {
	matches := videoIDRegex.FindStringSubmatch(url)
	if len(matches) < 2 {
		return nil, fmt.Errorf("invalid Kick VOD URL: %s", url)
	}

	videoID := matches[1]

	info, err := c.api.FetchVideoInfo(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("fetch Kick video %s: %w", videoID, err)
	}

	return &downloader.Video{
		Title: info.Title,
		URL:   info.SourceURL,
	}, nil
}

// Download fetches and writes a Kick VOD to outputPath.
func (c *Client) Download(ctx context.Context, video *downloader.Video, quality, outputPath string) error {
	if err := hls.Download(ctx, video.URL, quality, outputPath, c.api.FetchPlaylist, hls.FetchOptions{
		Concurrency: c.concurrency,
		Fetch:       c.api.FetchSegment,
	}); err != nil {
		return fmt.Errorf("download Kick segments: %w", err)
	}
	return nil
}
