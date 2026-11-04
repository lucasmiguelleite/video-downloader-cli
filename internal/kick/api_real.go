package kick

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// RealAPI calls Kick's production HTTP endpoints.
type RealAPI struct {
	client *http.Client
}

// NewRealAPI creates a production Kick HTTP client.
func NewRealAPI() *RealAPI {
	return &RealAPI{
		client: &http.Client{Timeout: 30 * time.Second},
	}
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

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
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

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("kick playlist fetch returned HTTP %d for %s", resp.StatusCode, m3u8URL)
	}

	data, err := io.ReadAll(resp.Body)
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

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("kick segment fetch returned HTTP %d for %s", resp.StatusCode, url)
	}

	return io.ReadAll(resp.Body)
}
