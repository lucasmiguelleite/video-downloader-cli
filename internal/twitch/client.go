package twitch

import (
	"fmt"
	"io"
	"regexp"
	"sync"

	"youtube-downloader/internal/downloader"
	"youtube-downloader/internal/hls"

	"github.com/schollz/progressbar/v3"
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

func (c *Client) GetVideo(url string) (*downloader.Video, error) {
	vodID := extractVodID(url)
	if vodID == "" {
		return nil, fmt.Errorf("invalid Twitch VOD URL: %s", url)
	}

	info, err := c.api.FetchVideoInfo(vodID)
	if err != nil {
		return nil, err
	}

	return &downloader.Video{
		Title: info.Title,
		URL:   vodID,
	}, nil
}

func (c *Client) Download(video *downloader.Video, quality string, w io.Writer) error {
	token, err := c.api.FetchAccessToken(video.URL)
	if err != nil {
		return err
	}

	usherURL := BuildUsherURL(video.URL, token)

	playlist, err := c.api.FetchPlaylist(usherURL)
	if err != nil {
		return err
	}

	segmentURL := usherURL

	if hls.IsMasterPlaylist(playlist) {
		segmentURL = hls.SelectVariant(usherURL, playlist, quality)
		if segmentURL == "" {
			return fmt.Errorf("no matching variant found for quality %s", quality)
		}

		playlist, err = c.api.FetchPlaylist(segmentURL)
		if err != nil {
			return err
		}
	}

	segments := hls.ParseSegments(segmentURL, playlist)
	if len(segments) == 0 {
		return fmt.Errorf("no segments found in playlist")
	}

	bar := progressbar.DefaultBytes(-1, "Downloading")

	for batchStart := 0; batchStart < len(segments); batchStart += c.concurrency {
		batchEnd := batchStart + c.concurrency
		if batchEnd > len(segments) {
			batchEnd = len(segments)
		}
		batch := segments[batchStart:batchEnd]

		results := make([][]byte, len(batch))
		var batchErr error
		var mu sync.Mutex
		var wg sync.WaitGroup

		for i, segURL := range batch {
			wg.Add(1)
			go func(idx int, url string) {
				defer wg.Done()
				data, err := c.api.FetchSegment(url)
				if err != nil {
					mu.Lock()
					if batchErr == nil {
						batchErr = fmt.Errorf("failed to download segment: %w", err)
					}
					mu.Unlock()
					return
				}
				results[idx] = data
			}(i, segURL)
		}

		wg.Wait()

		if batchErr != nil {
			return batchErr
		}

		for _, data := range results {
			w.Write(data)
			bar.Add(len(data))
		}
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
