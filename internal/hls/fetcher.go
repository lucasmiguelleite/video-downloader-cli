package hls

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
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
	// RetryConfig overrides the retry timing for this operation when non-nil.
	RetryConfig *downloader.RetryConfig
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
	opts, err := normalizedFetchOptions(opts)
	if err != nil {
		return err
	}
	state, err := openFetchState(segments, opts.Log, outputPath)
	if err != nil {
		return err
	}
	defer state.output.Close()
	if err := state.download(ctx, segments, opts, outputPath); err != nil {
		return err
	}
	if err := state.manifest.Delete(state.sidecarPath); err != nil {
		return fmt.Errorf("delete resume sidecar: %w", err)
	}
	return nil
}

type fetchState struct {
	output      *os.File
	manifest    *Manifest
	sidecarPath string
	completed   int
}

func normalizedFetchOptions(opts FetchOptions) (FetchOptions, error) {
	if opts.Fetch == nil {
		return opts, fmt.Errorf("hls: nil fetch function")
	}
	if opts.Concurrency < 1 {
		opts.Concurrency = 1
	}
	if opts.RetryAttempts < 1 {
		opts.RetryAttempts = 3
	}
	if opts.Log == nil {
		opts.Log = slog.Default()
	}
	return opts, nil
}

func openFetchState(segments []string, log *slog.Logger, outputPath string) (*fetchState, error) {
	sidecarPath := outputPath + ".part"
	manifest, err := LoadManifest(sidecarPath)
	if err != nil {
		return nil, fmt.Errorf("read resume sidecar: %w", err)
	}
	if manifest != nil && manifest.Matches(segments) && manifest.Completed > 0 && manifest.ValidOutput(outputPath) {
		return openResumeState(manifest, sidecarPath, outputPath, len(segments), log)
	}
	return openFreshState(segments, sidecarPath, outputPath)
}

func openResumeState(manifest *Manifest, sidecarPath, outputPath string, total int, log *slog.Logger) (*fetchState, error) {
	completed := min(manifest.Completed, total)
	output, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open output for resume: %w", err)
	}
	log.Info("resuming download", "skipped", completed, "remaining", total-completed)
	return &fetchState{output: output, manifest: manifest, sidecarPath: sidecarPath, completed: completed}, nil
}

func openFreshState(segments []string, sidecarPath, outputPath string) (*fetchState, error) {
	output, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("create output: %w", err)
	}
	manifest := &Manifest{Segments: segments}
	if err := manifest.Save(sidecarPath); err != nil {
		_ = output.Close()
		return nil, fmt.Errorf("write resume sidecar: %w", err)
	}
	return &fetchState{output: output, manifest: manifest, sidecarPath: sidecarPath}, nil
}

func (s *fetchState) download(ctx context.Context, segments []string, opts FetchOptions, outputPath string) error {
	bar := progressbar.DefaultBytes(-1, "Downloading")
	for start := s.completed; start < len(segments); start += opts.Concurrency {
		if err := ctx.Err(); err != nil {
			return err
		}
		end := min(start+opts.Concurrency, len(segments))
		paths, err := fetchAndStageBatch(ctx, segments[start:end], opts, outputPath)
		if err != nil {
			return err
		}
		if err := s.appendBatch(paths, bar); err != nil {
			cleanupSegmentFiles(paths)
			return err
		}
		cleanupSegmentFiles(paths)
		if err := s.recordBatch(end - start); err != nil {
			return err
		}
	}
	return nil
}

func fetchAndStageBatch(ctx context.Context, segments []string, opts FetchOptions, outputPath string) ([]string, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	paths := make([]string, len(segments))
	var firstErr error
	var mu sync.Mutex
	var wg sync.WaitGroup
	for i, segmentURL := range segments {
		wg.Add(1)
		go func(index int, rawURL string) {
			defer wg.Done()
			data, err := fetchSegmentWithRetry(ctx, rawURL, opts)
			if err == nil {
				paths[index], err = writeSegmentFile(outputPath, data)
			}
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("fetch and stage segment %s: %w", rawURL, err)
					cancel()
				}
				mu.Unlock()
			}
		}(i, segmentURL)
	}
	wg.Wait()
	if firstErr != nil {
		cleanupSegmentFiles(paths)
		return nil, firstErr
	}
	return paths, nil
}

func fetchSegmentWithRetry(ctx context.Context, rawURL string, opts FetchOptions) ([]byte, error) {
	var data []byte
	retry := func() error {
		var err error
		data, err = opts.Fetch(ctx, rawURL)
		return err
	}
	if opts.RetryConfig != nil {
		err := downloader.WithRetryConfig(ctx, opts.RetryAttempts, *opts.RetryConfig, retry)
		return data, err
	}
	err := downloader.WithRetry(ctx, opts.RetryAttempts, retry)
	return data, err
}

func (s *fetchState) appendBatch(paths []string, bar *progressbar.ProgressBar) error {
	for _, path := range paths {
		written, err := appendSegmentFile(s.output, path)
		if err != nil {
			return err
		}
		_ = bar.Add64(written)
	}
	if err := s.output.Sync(); err != nil {
		return fmt.Errorf("sync output: %w", err)
	}
	return nil
}

func (s *fetchState) recordBatch(count int) error {
	s.completed += count
	s.manifest.Completed = s.completed
	info, err := s.output.Stat()
	if err != nil {
		return fmt.Errorf("stat output: %w", err)
	}
	s.manifest.Bytes = info.Size()
	if err := s.manifest.Save(s.sidecarPath); err != nil {
		return fmt.Errorf("update resume sidecar: %w", err)
	}
	return nil
}

func cleanupSegmentFiles(paths []string) {
	for _, path := range paths {
		if path != "" {
			_ = os.Remove(path)
		}
	}
}

func writeSegmentFile(outputPath string, data []byte) (string, error) {
	f, err := os.CreateTemp(filepath.Dir(outputPath), ".segment-*")
	if err != nil {
		return "", err
	}
	path := f.Name()
	if _, err := f.Write(data); err != nil {
		f.Close()
		os.Remove(path)
		return "", err
	}
	if err := f.Close(); err != nil {
		os.Remove(path)
		return "", err
	}
	return path, nil
}

func appendSegmentFile(output *os.File, path string) (int64, error) {
	segment, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("open staged segment %q: %w", path, err)
	}
	written, copyErr := io.Copy(output, segment)
	closeErr := segment.Close()
	removeErr := os.Remove(path)
	if copyErr != nil {
		return 0, fmt.Errorf("append staged segment %q: %w", path, copyErr)
	}
	if closeErr != nil {
		return 0, fmt.Errorf("close staged segment %q: %w", path, closeErr)
	}
	if removeErr != nil {
		return 0, fmt.Errorf("remove staged segment %q: %w", path, removeErr)
	}
	return written, nil
}
