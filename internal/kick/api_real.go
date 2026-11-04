package kick

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"youtube-downloader/internal/httpbody"
)

const maxResponseBytes = 32 << 20

// RealAPI calls Kick's production HTTP endpoints.
type RealAPI struct {
	client *http.Client
}

// NewRealAPI creates a production Kick HTTP client.
func NewRealAPI() *RealAPI {
	return NewRealAPIWithHTTPClient(&http.Client{Timeout: 30 * time.Second})
}

// NewRealAPIWithHTTPClient creates a Kick API adapter using client.
func NewRealAPIWithHTTPClient(client *http.Client) *RealAPI {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	return &RealAPI{client: client}
}

func (api *RealAPI) FetchVideoInfo(ctx context.Context, videoID string) (*VideoInfo, error) {
	url := fmt.Sprintf("https://kick.com/api/v1/video/%s", videoID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)")
	req.Header.Set("Accept", "application/json")

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("kick API returned status %d", resp.StatusCode)
	}

	var result struct {
		Source     string `json:"source"`
		Livestream struct {
			SessionTitle string `json:"session_title"`
		} `json:"livestream"`
	}

	body, err := httpbody.Read(resp.Body, maxResponseBytes)
	if err != nil {
		return nil, fmt.Errorf("read Kick video response: %w", err)
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode Kick video response: %w", err)
	}

	title := result.Livestream.SessionTitle
	if title == "" {
		title = videoID
	}

	return &VideoInfo{
		Title:     title,
		SourceURL: result.Source,
	}, nil
}

func (api *RealAPI) FetchPlaylist(ctx context.Context, m3u8URL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m3u8URL, nil)
	if err != nil {
		return "", err
	}
	resp, err := api.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("kick playlist fetch returned HTTP %d for %s", resp.StatusCode, m3u8URL)
	}

	data, err := httpbody.Read(resp.Body, maxResponseBytes)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (api *RealAPI) FetchSegment(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := api.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("kick segment fetch returned HTTP %d for %s", resp.StatusCode, url)
	}

	return httpbody.Read(resp.Body, maxResponseBytes)
}
