package hls

import (
	"context"
	"fmt"
)

// PlaylistFetcher retrieves an HLS playlist by URL.
type PlaylistFetcher func(ctx context.Context, url string) (string, error)

// Download resolves an HLS playlist, selects quality, and writes its segments.
func Download(ctx context.Context, playlistURL, quality, outputPath string, fetchPlaylist PlaylistFetcher, opts FetchOptions) error {
	playlist, err := fetchPlaylist(ctx, playlistURL)
	if err != nil {
		return fmt.Errorf("fetch playlist %s: %w", playlistURL, err)
	}
	segmentURL := playlistURL
	if IsMasterPlaylist(playlist) {
		segmentURL, err = SelectVariant(playlistURL, playlist, quality)
		if err != nil {
			return fmt.Errorf("select stream variant: %w", err)
		}
		if segmentURL == "" {
			return fmt.Errorf("no matching variant found for quality %s", quality)
		}
		playlist, err = fetchPlaylist(ctx, segmentURL)
		if err != nil {
			return fmt.Errorf("fetch media playlist %s: %w", segmentURL, err)
		}
	}
	segments, err := ParseSegments(segmentURL, playlist)
	if err != nil {
		return fmt.Errorf("parse media playlist: %w", err)
	}
	if len(segments) == 0 {
		return fmt.Errorf("no segments found in playlist")
	}
	return FetchSegments(ctx, segments, opts, outputPath)
}
