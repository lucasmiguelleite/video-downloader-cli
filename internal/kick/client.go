package kick

import (
	"fmt"
	"io"
	"regexp"

	"youtube-downloader/internal/downloader"

	"github.com/schollz/progressbar/v3"
)

var videoIDRegex = regexp.MustCompile(`kick\.com/\w+/videos/([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})`)

type Client struct {
	api KickAPI
}

func NewClient(api KickAPI) *Client {
	return &Client{api: api}
}

func (c *Client) GetVideo(url string) (*downloader.Video, error) {
	matches := videoIDRegex.FindStringSubmatch(url)
	if len(matches) < 2 {
		return nil, fmt.Errorf("invalid Kick VOD URL: %s", url)
	}

	videoID := matches[1]

	info, err := c.api.FetchVideoInfo(videoID)
	if err != nil {
		return nil, err
	}

	return &downloader.Video{
		Title: info.Title,
		URL:   info.SourceURL,
	}, nil
}

func (c *Client) Download(video *downloader.Video, quality string, w io.Writer) error {
	playlist, err := c.api.FetchPlaylist(video.URL)
	if err != nil {
		return err
	}

	segmentURL := video.URL

	if IsMasterPlaylist(playlist) {
		segmentURL = SelectVariant(video.URL, playlist, quality)
		if segmentURL == "" {
			return fmt.Errorf("no matching variant found for quality %s", quality)
		}

		playlist, err = c.api.FetchPlaylist(segmentURL)
		if err != nil {
			return err
		}
	}

	segments := ParseSegments(segmentURL, playlist)
	if len(segments) == 0 {
		return fmt.Errorf("no segments found in playlist")
	}

	bar := progressbar.DefaultBytes(
		int64(len(segments)),
		"Downloading",
	)

	for i, segURL := range segments {
		data, err := c.api.FetchSegment(segURL)
		if err != nil {
			return fmt.Errorf("failed to download segment %d: %w", i, err)
		}
		w.Write(data)
		bar.Add(1)
	}

	return nil
}
