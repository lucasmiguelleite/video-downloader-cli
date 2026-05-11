package kick

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type RealAPI struct {
	client *http.Client
}

func NewRealAPI() *RealAPI {
	return &RealAPI{
		client: &http.Client{},
	}
}

func (api *RealAPI) FetchVideoInfo(videoID string) (*VideoInfo, error) {
	url := fmt.Sprintf("https://kick.com/api/v1/video/%s", videoID)

	req, err := http.NewRequest("GET", url, nil)
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
		Source string `json:"source"`
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

func (api *RealAPI) FetchPlaylist(m3u8URL string) (string, error) {
	resp, err := api.client.Get(m3u8URL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (api *RealAPI) FetchSegment(url string) ([]byte, error) {
	resp, err := api.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
