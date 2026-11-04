package app

import (
	"context"
	"errors"
	"log/slog"
	"path/filepath"

	"youtube-downloader/internal/downloader"
	"youtube-downloader/internal/fs"
)

// DownloadUseCase coordinates video resolution, download, and optional remux.
type DownloadUseCase struct {
	service    downloader.Service
	dir        string
	needsRemux bool
	log        *slog.Logger
}

// NewDownloadUseCase creates a download use case using the default directory.
func NewDownloadUseCase(s downloader.Service) *DownloadUseCase {
	return &DownloadUseCase{service: s, log: slog.Default()}
}

// NewDownloadUseCaseWithDir creates a download use case writing to dir.
func NewDownloadUseCaseWithDir(s downloader.Service, dir string) *DownloadUseCase {
	return &DownloadUseCase{service: s, dir: dir, log: slog.Default()}
}

// NewDownloadUseCaseWithRemux creates a use case that remuxes HLS output.
func NewDownloadUseCaseWithRemux(s downloader.Service, dir string) *DownloadUseCase {
	return &DownloadUseCase{service: s, dir: dir, needsRemux: true, log: slog.Default()}
}

// WithLogger attaches a structured logger. Returns the use case for chaining.
// A nil logger resets to slog.Default().
func (uc *DownloadUseCase) WithLogger(log *slog.Logger) *DownloadUseCase {
	if log == nil {
		log = slog.Default()
	}
	uc.log = log
	return uc
}

// Execute resolves and downloads url at quality, returning the saved path.
func (uc *DownloadUseCase) Execute(ctx context.Context, url, quality string) (string, error) {
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

	uc.log.Info("download starting", "title", video.Title, "dir", dir, "quality", quality)

	if err := uc.service.Download(ctx, video, quality, path); err != nil {
		// Graceful shutdown: on cancellation, keep the partial file so the
		// user can re-run to resume. Report the path rather than an error.
		if errors.Is(err, context.Canceled) {
			uc.log.Info("download interrupted, partial file kept — re-run to resume", "path", path)
			return path, nil
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
