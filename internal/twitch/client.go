package twitch

import (
	"context"
	"fmt"
	"regexp"

	"youtube-downloader/internal/downloader"
	"youtube-downloader/internal/hls"
)

var vodIDRegex = regexp.MustCompile(`twitch\.tv/(?:[^/]+/v(?:ideo)?|videos)/(\d+)|twitch\.tv/.*[?&]video=v?(\d+)`)

type Client struct {
	api         TwitchAPI
	concurrency int
}

func NewClient(api TwitchAPI) *Client {
	return &Client{api: api, concurrency: 20}
}

func NewClientWithConcurrency(api TwitchAPI, concurrency int) *Client {
	if concurrency < 1 {
		concurrency = 1
	}
	return &Client{api: api, concurrency: concurrency}
}

func (c *Client) GetVideo(ctx context.Context, url string) (*downloader.Video, error) {
	vodID := extractVodID(url)
	if vodID == "" {
		return nil, fmt.Errorf("invalid Twitch VOD URL: %s", url)
	}

	info, err := c.api.FetchVideoInfo(ctx, vodID)
	if err != nil {
		return nil, err
	}

	return &downloader.Video{
		Title: info.Title,
		URL:   vodID,
	}, nil
}

func (c *Client) Download(ctx context.Context, video *downloader.Video, quality, outputPath string) error {
	token, err := c.api.FetchAccessToken(ctx, video.URL)
	if err != nil {
		return err
	}

	usherURL := BuildUsherURL(video.URL, token)

	playlist, err := c.api.FetchPlaylist(ctx, usherURL)
	if err != nil {
		return err
	}

	segmentURL := usherURL

	if hls.IsMasterPlaylist(playlist) {
		segmentURL = hls.SelectVariant(usherURL, playlist, quality)
		if segmentURL == "" {
			return fmt.Errorf("no matching variant found for quality %s", quality)
		}

		playlist, err = c.api.FetchPlaylist(ctx, segmentURL)
		if err != nil {
			return err
		}
	}

	segments := hls.ParseSegments(segmentURL, playlist)
	if len(segments) == 0 {
		return fmt.Errorf("no segments found in playlist")
	}

	return hls.FetchSegments(ctx, segments, hls.FetchOptions{
		Concurrency: c.concurrency,
		Fetch:       c.api.FetchSegment,
	}, outputPath)
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
