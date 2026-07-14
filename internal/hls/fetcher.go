package hls

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"youtube-downloader/internal/downloader"

	"github.com/schollz/progressbar/v3"
)

// SegmentFetcher downloads a single HLS segment and returns its raw bytes.
type SegmentFetcher func(ctx context.Context, url string) ([]byte, error)

// FetchOptions configures FetchSegments.
type FetchOptions struct {
	// Concurrency caps parallel segment downloads (min 1).
	Concurrency int
	// Fetch is the per-segment downloader.
	Fetch SegmentFetcher
	// RetryAttempts is the max attempts per segment (min 1). 1 disables retry.
	RetryAttempts int
	// Log, if non-nil, is used for resume/debug messages.
	Log *slog.Logger
}

// FetchSegments downloads the given segment URLs with bounded concurrency,
// appending the results to outputPath in the original playlist order.
//
// Resume behaviour:
//   - If outputPath+".part" exists and its segment list matches, the run
//     resumes: completed segments are skipped, the output file is opened in
//     append mode, and only the remaining segments are fetched.
//   - Otherwise (no sidecar, or stale segment list), the run starts fresh:
//     the output file is truncated and all segments are fetched.
//
// After each successfully written batch, the sidecar is updated with the new
// completed-count, so an interruption at any point leaves a resumable state.
// On full completion, the sidecar is deleted.
func FetchSegments(ctx context.Context, segments []string, opts FetchOptions, outputPath string) error {
	if len(segments) == 0 {
		return nil
	}
	if opts.Concurrency < 1 {
		opts.Concurrency = 1
	}
	if opts.RetryAttempts < 1 {
		opts.RetryAttempts = 3 // sensible default for transient CDN errors
	}
	if opts.Fetch == nil {
		return fmt.Errorf("hls: nil fetch function")
	}
	log := opts.Log
	if log == nil {
		log = slog.Default()
	}

	sidecarPath := outputPath + ".part"

	// Decide whether to resume or start over.
	startIdx := 0
	existing, err := LoadManifest(sidecarPath)
	if err != nil {
		return fmt.Errorf("read resume sidecar: %w", err)
	}
	if existing != nil && existing.Matches(segments) && existing.Completed > 0 {
		startIdx = existing.Completed
		if startIdx > len(segments) {
			startIdx = len(segments)
		}
		log.Info("resuming download", "skipped", startIdx, "remaining", len(segments)-startIdx)
	} else {
		// Fresh start — drop any stale sidecar; the file will be truncated below.
		existing = nil
	}

	// Open the output file in the right mode and keep the manifest in sync.
	var manifest *Manifest
	var f *os.File
	if existing != nil {
		f, err = os.OpenFile(outputPath, os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return fmt.Errorf("open output for resume: %w", err)
		}
		manifest = existing
	} else {
		f, err = os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("create output: %w", err)
		}
		manifest = &Manifest{Segments: segments, Completed: 0}
		// Persist the initial manifest before doing any work, so an
		// interruption at any point has a sidecar to resume from.
		if err := manifest.Save(sidecarPath); err != nil {
			f.Close()
			return fmt.Errorf("write resume sidecar: %w", err)
		}
	}
	defer f.Close()

	bar := progressbar.DefaultBytes(-1, "Downloading")

	// fetchOne applies retry around the user-provided fetcher.
	fetchOne := func(ctx context.Context, url string) ([]byte, error) {
		var data []byte
		err := downloader.WithRetry(ctx, opts.RetryAttempts, func() error {
			d, e := opts.Fetch(ctx, url)
			if e != nil {
				return e
			}
			data = d
			return nil
		})
		return data, err
	}

	completed := startIdx
	for batchStart := startIdx; batchStart < len(segments); batchStart += opts.Concurrency {
		if err := ctx.Err(); err != nil {
			return err
		}

		batchEnd := batchStart + opts.Concurrency
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
				data, err := fetchOne(ctx, url)
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

		// Append this batch in playlist order.
		for _, data := range results {
			if _, err := f.Write(data); err != nil {
				return err
			}
			bar.Add(len(data))
		}
		if err := f.Sync(); err != nil {
			return err
		}

		// Update sidecar so a crash after this point is resumable.
		completed += len(batch)
		manifest.Completed = completed
		if err := manifest.Save(sidecarPath); err != nil {
			return fmt.Errorf("update resume sidecar: %w", err)
		}
	}

	// All done — clean up the sidecar.
	if err := manifest.Delete(sidecarPath); err != nil {
		return fmt.Errorf("delete resume sidecar: %w", err)
	}
	return nil
}
