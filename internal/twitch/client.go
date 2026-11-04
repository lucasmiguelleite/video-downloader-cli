package twitch

import (
	"context"
	"fmt"
	"regexp"

	"youtube-downloader/internal/downloader"
	"youtube-downloader/internal/hls"
)

var vodIDRegex = regexp.MustCompile(`twitch\.tv/(?:[^/]+/v(?:ideo)?|videos)/(\d+)|twitch\.tv/.*[?&]video=v?(\d+)`)

// Client downloads Twitch VODs through a TwitchAPI.
type Client struct {
	api         TwitchAPI
	concurrency int
}

// NewClient creates a Twitch downloader with the default concurrency.
func NewClient(api TwitchAPI) *Client {
	return &Client{api: api, concurrency: 20}
}

// NewClientWithConcurrency creates a Twitch downloader with bounded concurrency.
func NewClientWithConcurrency(api TwitchAPI, concurrency int) *Client {
	if concurrency < 1 {
		concurrency = 1
	}
	return &Client{api: api, concurrency: concurrency}
}

// GetVideo resolves Twitch VOD metadata from url.
func (c *Client) GetVideo(ctx context.Context, url string) (*downloader.Video, error) {
	vodID := extractVodID(url)
	if vodID == "" {
		return nil, fmt.Errorf("invalid Twitch VOD URL: %s", url)
	}

	info, err := c.api.FetchVideoInfo(ctx, vodID)
	if err != nil {
		return nil, fmt.Errorf("fetch Twitch video %s: %w", vodID, err)
	}

	return &downloader.Video{
		Title: info.Title,
		URL:   vodID,
	}, nil
}

// Download fetches and writes a Twitch VOD to outputPath.
func (c *Client) Download(ctx context.Context, video *downloader.Video, quality, outputPath string) error {
	token, err := c.api.FetchAccessToken(ctx, video.URL)
	if err != nil {
		return fmt.Errorf("fetch Twitch access token for %s: %w", video.URL, err)
	}

	usherURL := BuildUsherURL(video.URL, token)

	if err := hls.Download(ctx, usherURL, quality, outputPath, c.api.FetchPlaylist, hls.FetchOptions{
		Concurrency: c.concurrency,
		Fetch:       c.api.FetchSegment,
	}); err != nil {
		return fmt.Errorf("download Twitch segments: %w", err)
	}
	return nil
}

func extractVodID(url string) string {
	matches := vodIDRegex.FindStringSubmatch(url)
	for i := 1; i < len(matches); i++ {
		if matches[i] != "" {
			return matches[i]
		}
	}
	return ""
}
