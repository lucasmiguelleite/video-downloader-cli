package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"

	"youtube-downloader/internal/downloader"
	"youtube-downloader/internal/fs"
)

// ErrDownloadInterrupted identifies a download stopped by context cancellation.
var ErrDownloadInterrupted = errors.New("download interrupted")

// DownloadUseCase coordinates video resolution, download, and optional remux.
type DownloadUseCase struct {
	service    downloader.Service
	dir        string
	needsRemux bool
	log        *slog.Logger
}

// DownloadOptions configures a DownloadUseCase.
type DownloadOptions struct {
	// Dir is the output directory. Empty uses the user's Downloads directory.
	Dir string
	// NeedsRemux enables a final ffmpeg stream-copy remux.
	NeedsRemux bool
	// Logger receives lifecycle events. Nil uses slog.Default().
	Logger *slog.Logger
}

// NewDownloadUseCaseWithOptions creates a configured download use case.
func NewDownloadUseCaseWithOptions(s downloader.Service, opts DownloadOptions) *DownloadUseCase {
	log := opts.Logger
	if log == nil {
		log = slog.Default()
	}
	return &DownloadUseCase{service: s, dir: opts.Dir, needsRemux: opts.NeedsRemux, log: log}
}

// Execute resolves and downloads url at quality, returning the saved path.
func (uc *DownloadUseCase) Execute(ctx context.Context, url, quality string) (resultPath string, resultErr error) {
	video, err := uc.service.GetVideo(ctx, url)
	if err != nil {
		uc.log.Error("failed to resolve video", "url", url, "err", err)
		return "", err
	}

	dir := uc.dir
	if dir == "" {
		dir, err = fs.DefaultDownloadDir()
		if err != nil {
			return "", err
		}
	}

	// The provider downloads directly into outputPath, which doubles as the
	// resume anchor (a .part sidecar lives next to it). Make sure the dir
	// exists and hand the provider the absolute path.
	if err := fs.MkdirAll(dir); err != nil {
		return "", err
	}

	filename := fs.SafeFilename(video.Title, ".mp4")
	path := filepath.Join(dir, filename)
	reservation, err := fs.ReservePath(path, uc.needsRemux)
	if err != nil {
		return "", err
	}
	path = reservation.Path
	defer func() {
		if err := reservation.Release(); err != nil && resultErr == nil {
			resultErr = fmt.Errorf("release output reservation for %q: %w", path, err)
		}
	}()

	uc.log.Info("download starting", "title", video.Title, "dir", dir, "quality", quality)

	if err := uc.service.Download(ctx, video, quality, path); err != nil {
		if errors.Is(err, context.Canceled) {
			uc.log.Info("download interrupted, partial file kept — re-run to resume", "path", path)
			return path, fmt.Errorf("%w: %w", ErrDownloadInterrupted, err)
		}
		uc.log.Error("download failed", "title", video.Title, "err", err)
		return "", err
	}

	if uc.needsRemux {
		uc.log.Info("remuxing to mp4", "path", path)
		if err := fs.RemuxToMP4(ctx, path); err != nil {
			return "", err
		}
	}

	uc.log.Info("download complete", "path", path)
	return path, nil
}
