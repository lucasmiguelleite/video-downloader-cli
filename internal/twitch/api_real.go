package twitch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"youtube-downloader/internal/httpbody"
)

const clientID = "ue6666qo983tsx6so1t0vnawi233wa"

const maxResponseBytes = 32 << 20

// RealAPI calls Twitch's production HTTP endpoints.
type RealAPI struct {
	client *http.Client
}

// NewRealAPI creates a production Twitch HTTP client.
func NewRealAPI() *RealAPI {
	return NewRealAPIWithHTTPClient(&http.Client{Timeout: 30 * time.Second})
}

// NewRealAPIWithHTTPClient creates a Twitch API adapter using client.
func NewRealAPIWithHTTPClient(client *http.Client) *RealAPI {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	return &RealAPI{client: client}
}

func (api *RealAPI) FetchVideoInfo(ctx context.Context, videoID string) (*VideoInfo, error) {
	query := fmt.Sprintf(`{
		video(id: "%s") {
			title
		}
	}`, videoID)

	data, err := api.graphql(ctx, query)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data struct {
			Video struct {
				Title string `json:"title"`
			} `json:"video"`
		} `json:"data"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	title := result.Data.Video.Title
	if title == "" {
		title = videoID
	}

	return &VideoInfo{Title: title}, nil
}

func (api *RealAPI) FetchAccessToken(ctx context.Context, videoID string) (*AccessToken, error) {
	query := fmt.Sprintf(`{
		videoPlaybackAccessToken(id: "%s", params: {platform: "web", playerBackend: "mediaplayer", playerType: "site"}) {
			value
			signature
		}
	}`, videoID)

	data, err := api.graphql(ctx, query)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data struct {
			Token struct {
				Value     string `json:"value"`
				Signature string `json:"signature"`
			} `json:"videoPlaybackAccessToken"`
		} `json:"data"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return &AccessToken{
		Value:     result.Data.Token.Value,
		Signature: result.Data.Token.Signature,
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
		return "", fmt.Errorf("twitch playlist fetch returned HTTP %d for %s", resp.StatusCode, m3u8URL)
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
		return nil, fmt.Errorf("twitch segment fetch returned HTTP %d for %s", resp.StatusCode, url)
	}

	return httpbody.Read(resp.Body, maxResponseBytes)
}

func (api *RealAPI) graphql(ctx context.Context, query string) ([]byte, error) {
	body, err := json.Marshal(map[string]string{"query": query})
	if err != nil {
		return nil, fmt.Errorf("encode Twitch GraphQL query: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://gql.twitch.tv/gql", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-ID", clientID)

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("twitch GraphQL returned status %d", resp.StatusCode)
	}

	return httpbody.Read(resp.Body, maxResponseBytes)
}

// BuildUsherURL returns the authenticated Twitch VOD playlist URL.
func BuildUsherURL(videoID string, token *AccessToken) string {
	params := url.Values{}
	params.Set("sig", token.Signature)
	params.Set("token", token.Value)
	params.Set("allow_source", "true")
	params.Set("allow_audio_only", "true")

	return fmt.Sprintf("https://usher.ttvnw.net/vod/%s.m3u8?%s", videoID, params.Encode())
}
