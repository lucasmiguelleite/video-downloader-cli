package twitch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const clientID = "ue6666qo983tsx6so1t0vnawi233wa"

type RealAPI struct {
	client *http.Client
}

func NewRealAPI() *RealAPI {
	return &RealAPI{client: &http.Client{}}
}

func (api *RealAPI) FetchVideoInfo(videoID string) (*VideoInfo, error) {
	query := fmt.Sprintf(`{
		video(id: "%s") {
			title
		}
	}`, videoID)

	data, err := api.graphql(query)
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

func (api *RealAPI) FetchAccessToken(videoID string) (*AccessToken, error) {
	query := fmt.Sprintf(`{
		videoPlaybackAccessToken(id: "%s", params: {platform: "web", playerBackend: "mediaplayer", playerType: "site"}) {
			value
			signature
		}
	}`, videoID)

	data, err := api.graphql(query)
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

func (api *RealAPI) graphql(query string) ([]byte, error) {
	body, _ := json.Marshal(map[string]string{"query": query})

	req, err := http.NewRequest("POST", "https://gql.twitch.tv/gql", bytes.NewReader(body))
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

	return io.ReadAll(resp.Body)
}

func BuildUsherURL(videoID string, token *AccessToken) string {
	params := url.Values{}
	params.Set("sig", token.Signature)
	params.Set("token", token.Value)
	params.Set("allow_source", "true")
	params.Set("allow_audio_only", "true")

	return fmt.Sprintf("https://usher.ttvnw.net/vod/%s.m3u8?%s", videoID, params.Encode())
}
