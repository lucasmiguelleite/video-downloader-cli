package kick

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"youtube-downloader/internal/downloader"
	"youtube-downloader/internal/hls"
)

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
	videoID, err := extractVideoID(url)
	if err != nil {
		return nil, err
	}

	info, err := c.api.FetchVideoInfo(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("fetch Kick video %s: %w", videoID, err)
	}

	return &downloader.Video{
		Title: info.Title,
		URL:   info.SourceURL,
	}, nil
}

func extractVideoID(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme != "https" || parsed.RawPath != "" {
		return "", fmt.Errorf("invalid Kick VOD URL %q; expected an unescaped HTTPS URL", rawURL)
	}
	host := strings.ToLower(parsed.Hostname())
	if host != "kick.com" && host != "www.kick.com" {
		return "", fmt.Errorf("invalid Kick host %q in %s", host, rawURL)
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) != 3 || parts[0] == "" || parts[1] != "videos" || !validUUID(parts[2]) {
		return "", fmt.Errorf("invalid Kick VOD path %q; expected /channel/videos/UUID", parsed.Path)
	}
	return strings.ToLower(parts[2]), nil
}

func validUUID(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' {
		return false
	}
	for i, r := range value {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			continue
		}
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
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
