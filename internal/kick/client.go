package kick

import (
	"context"
	"fmt"
	"regexp"

	"youtube-downloader/internal/downloader"
	"youtube-downloader/internal/hls"
)

var videoIDRegex = regexp.MustCompile(`kick\.com/\w+/videos/([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})`)

type Client struct {
	api         KickAPI
	concurrency int
}

func NewClient(api KickAPI) *Client {
	return &Client{api: api, concurrency: 20}
}

func NewClientWithConcurrency(api KickAPI, concurrency int) *Client {
	if concurrency < 1 {
		concurrency = 1
	}
	return &Client{api: api, concurrency: concurrency}
}

func (c *Client) GetVideo(ctx context.Context, url string) (*downloader.Video, error) {
	matches := videoIDRegex.FindStringSubmatch(url)
	if len(matches) < 2 {
		return nil, fmt.Errorf("invalid Kick VOD URL: %s", url)
	}

	videoID := matches[1]

	info, err := c.api.FetchVideoInfo(ctx, videoID)
	if err != nil {
		return nil, err
	}

	return &downloader.Video{
		Title: info.Title,
		URL:   info.SourceURL,
	}, nil
}

func (c *Client) Download(ctx context.Context, video *downloader.Video, quality, outputPath string) error {
	playlist, err := c.api.FetchPlaylist(ctx, video.URL)
	if err != nil {
		return err
	}

	segmentURL := video.URL

	if hls.IsMasterPlaylist(playlist) {
		segmentURL = hls.SelectVariant(video.URL, playlist, quality)
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
