package twitch

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"youtube-downloader/internal/downloader"
	"youtube-downloader/internal/hls"
)

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
	vodID, err := extractVodID(url)
	if err != nil {
		return nil, err
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

func extractVodID(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme != "https" || parsed.RawPath != "" {
		return "", fmt.Errorf("invalid Twitch VOD URL %q; expected an unescaped HTTPS URL", rawURL)
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "player.twitch.tv" {
		return twitchPlayerID(parsed, rawURL)
	}
	if host != "twitch.tv" && host != "www.twitch.tv" {
		return "", fmt.Errorf("invalid Twitch host %q in %s", host, rawURL)
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	var id string
	if len(parts) == 2 && parts[0] == "videos" {
		id = parts[1]
	} else if len(parts) == 3 && (parts[1] == "v" || parts[1] == "video") {
		id = parts[2]
	}
	if !digitsOnly(id) {
		return "", fmt.Errorf("invalid Twitch VOD identifier in %s", rawURL)
	}
	return id, nil
}

func twitchPlayerID(parsed *url.URL, rawURL string) (string, error) {
	id := strings.TrimPrefix(parsed.Query().Get("video"), "v")
	if !digitsOnly(id) {
		return "", fmt.Errorf("invalid Twitch player video identifier in %s", rawURL)
	}
	return id, nil
}

func digitsOnly(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
