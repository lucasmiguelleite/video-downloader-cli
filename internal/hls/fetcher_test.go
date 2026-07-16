package hls

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"youtube-downloader/internal/downloader"
)

func writeFetch(data map[string][]byte) SegmentFetcher {
	return func(ctx context.Context, url string) ([]byte, error) {
		return data[url], nil
	}
}

func TestFetchSegments_WritesAllSegmentsInOrder(t *testing.T) {
	segments := []string{
		"https://cdn.example.com/seg1.ts",
		"https://cdn.example.com/seg2.ts",
		"https://cdn.example.com/seg3.ts",
		"https://cdn.example.com/seg4.ts",
	}
	data := map[string][]byte{
		"https://cdn.example.com/seg1.ts": []byte("data1"),
		"https://cdn.example.com/seg2.ts": []byte("data2"),
		"https://cdn.example.com/seg3.ts": []byte("data3"),
		"https://cdn.example.com/seg4.ts": []byte("data4"),
	}

	path := filepath.Join(t.TempDir(), "out.ts")
	err := FetchSegments(context.Background(), segments, FetchOptions{
		Concurrency: 2,
		Fetch:       writeFetch(data),
	}, path)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	got, _ := os.ReadFile(path)
	expected := "data1data2data3data4"
	if string(got) != expected {
		t.Errorf("expected %q, got %q", expected, string(got))
	}
}

func TestFetchSegments_ReturnsFetchError(t *testing.T) {
	segments := []string{"https://cdn.example.com/seg1.ts", "https://cdn.example.com/bad.ts"}
	errSentinel := errors.New("boom")

	fetch := func(ctx context.Context, url string) ([]byte, error) {
		if url == "https://cdn.example.com/bad.ts" {
			return nil, errSentinel
		}
		return []byte("ok"), nil
	}

	path := filepath.Join(t.TempDir(), "out.ts")
	err := FetchSegments(context.Background(), segments, FetchOptions{
		Concurrency: 1,
		Fetch:       fetch,
	}, path)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errSentinel) {
		t.Errorf("expected wrapped sentinel error, got: %v", err)
	}
}

func TestFetchSegments_RespectsContextCancellation(t *testing.T) {
	segments := []string{"https://cdn.example.com/seg1.ts", "https://cdn.example.com/seg2.ts"}

	fetch := func(ctx context.Context, url string) ([]byte, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		return []byte("data"), nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	path := filepath.Join(t.TempDir(), "out.ts")
	err := FetchSegments(ctx, segments, FetchOptions{Concurrency: 1, Fetch: fetch}, path)

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got: %v", err)
	}
}

func TestFetchSegments_ConcurrencyOneIsSequential(t *testing.T) {
	segments := []string{
		"https://cdn.example.com/seg1.ts",
		"https://cdn.example.com/seg2.ts",
		"https://cdn.example.com/seg3.ts",
	}

	var inflight int32
	var maxInflight int32
	var mu sync.Mutex

	fetch := func(ctx context.Context, url string) ([]byte, error) {
		cur := atomic.AddInt32(&inflight, 1)
		mu.Lock()
		if cur > maxInflight {
			maxInflight = cur
		}
		mu.Unlock()
		time.Sleep(5 * time.Millisecond)
		atomic.AddInt32(&inflight, -1)
		return []byte("x"), nil
	}

	path := filepath.Join(t.TempDir(), "out.ts")
	err := FetchSegments(context.Background(), segments, FetchOptions{Concurrency: 1, Fetch: fetch}, path)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if maxInflight != 1 {
		t.Errorf("with concurrency=1 expected maxInflight=1, got %d", maxInflight)
	}
	got, _ := os.ReadFile(path)
	if len(got) != 3 {
		t.Errorf("expected 3 bytes written, got %q", string(got))
	}
}

func TestFetchSegments_ConcurrencyBoundedToN(t *testing.T) {
	segments := make([]string, 20)
	for i := range segments {
		segments[i] = "seg"
	}

	var inflight int32
	var maxInflight int32
	var mu sync.Mutex

	fetch := func(ctx context.Context, url string) ([]byte, error) {
		cur := atomic.AddInt32(&inflight, 1)
		mu.Lock()
		if cur > maxInflight {
			maxInflight = cur
		}
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
		atomic.AddInt32(&inflight, -1)
		return []byte("x"), nil
	}

	path := filepath.Join(t.TempDir(), "out.ts")
	if err := FetchSegments(context.Background(), segments, FetchOptions{Concurrency: 4, Fetch: fetch}, path); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if maxInflight > 4 {
		t.Errorf("expected maxInflight <= 4, got %d", maxInflight)
	}
}

func TestFetchSegments_EmptySegmentList(t *testing.T) {
	fetch := func(ctx context.Context, url string) ([]byte, error) {
		t.Error("fetch should not be called for empty segment list")
		return nil, nil
	}
	path := filepath.Join(t.TempDir(), "out.ts")
	if err := FetchSegments(context.Background(), nil, FetchOptions{Concurrency: 4, Fetch: fetch}, path); err != nil {
		t.Fatalf("expected no error for empty input, got: %v", err)
	}
	if got, _ := os.ReadFile(path); len(got) != 0 {
		t.Errorf("expected empty output, got %q", string(got))
	}
}

func TestFetchSegments_RetriesTransientSegmentFailure(t *testing.T) {
	// Speed up backoff for this test.
	orig := downloader.DefaultRetryConfig
	downloader.DefaultRetryConfig = downloader.RetryConfig{Base: time.Millisecond, Max: 5 * time.Millisecond, Jitter: 0}
	t.Cleanup(func() { downloader.DefaultRetryConfig = orig })

	segments := []string{
		"https://cdn.example.com/flaky.ts",
		"https://cdn.example.com/stable.ts",
	}
	fetch := newFlakyFetch(map[string][]byte{
		"https://cdn.example.com/flaky.ts":  []byte("flaky"),
		"https://cdn.example.com/stable.ts": []byte("stable"),
	}, map[string]int{
		"https://cdn.example.com/flaky.ts": 2,
	})

	path := filepath.Join(t.TempDir(), "out.ts")
	err := FetchSegments(context.Background(), segments, FetchOptions{Concurrency: 2, Fetch: fetch}, path)
	if err != nil {
		t.Fatalf("expected retry to recover, got: %v", err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != "flakystable" {
		t.Errorf("expected 'flakystable', got %q", string(got))
	}
}

// TestFetchSegments_FreshDownloadDeletesSidecar verifies that on a successful
// fresh download, the .part sidecar is removed.
func TestFetchSegments_FreshDownloadDeletesSidecar(t *testing.T) {
	segments := []string{"https://cdn.example.com/seg1.ts", "https://cdn.example.com/seg2.ts"}
	data := map[string][]byte{
		"https://cdn.example.com/seg1.ts": []byte("d1"),
		"https://cdn.example.com/seg2.ts": []byte("d2"),
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "out.ts")
	if err := FetchSegments(context.Background(), segments, FetchOptions{Concurrency: 2, Fetch: writeFetch(data)}, path); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if _, err := os.Stat(path + ".part"); !os.IsNotExist(err) {
		t.Errorf("expected sidecar to be deleted, got stat err: %v", err)
	}
}

// TestFetchSegments_ResumeSkipsCompletedSegments verifies that when a valid
// sidecar exists, the already-completed segments are skipped and the run
// appends only the remaining ones.
func TestFetchSegments_ResumeSkipsCompletedSegments(t *testing.T) {
	segments := []string{
		"https://cdn.example.com/seg1.ts",
		"https://cdn.example.com/seg2.ts",
		"https://cdn.example.com/seg3.ts",
	}
	data := map[string][]byte{
		"https://cdn.example.com/seg1.ts": []byte("d1"),
		"https://cdn.example.com/seg2.ts": []byte("d2"),
		"https://cdn.example.com/seg3.ts": []byte("d3"),
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "out.ts")

	// Simulate a previous run that completed 2 of 3 segments: write the
	// first two segments to the output file and persist the sidecar.
	if err := os.WriteFile(path, []byte("d1d2"), 0o644); err != nil {
		t.Fatal(err)
	}
	sidecar := &Manifest{Segments: segments, Completed: 2}
	if err := sidecar.Save(path + ".part"); err != nil {
		t.Fatal(err)
	}

	// Track which segments get fetched.
	var fetched []string
	var mu sync.Mutex
	fetch := func(ctx context.Context, url string) ([]byte, error) {
		mu.Lock()
		fetched = append(fetched, url)
		mu.Unlock()
		return data[url], nil
	}

	if err := FetchSegments(context.Background(), segments, FetchOptions{Concurrency: 2, Fetch: fetch}, path); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Only seg3 should have been fetched.
	if len(fetched) != 1 || fetched[0] != "https://cdn.example.com/seg3.ts" {
		t.Errorf("expected only seg3 fetched, got %v", fetched)
	}

	got, _ := os.ReadFile(path)
	if string(got) != "d1d2d3" {
		t.Errorf("expected 'd1d2d3' (resumed append), got %q", string(got))
	}
	if _, err := os.Stat(path + ".part"); !os.IsNotExist(err) {
		t.Errorf("sidecar should be deleted after completion, got stat err: %v", err)
	}
}

// TestFetchSegments_StaleSidecarStartsOver verifies that if the sidecar's
// segment list no longer matches (e.g. VOD was reprocessed), the download
// starts fresh from segment 0.
func TestFetchSegments_StaleSidecarStartsOver(t *testing.T) {
	segments := []string{"https://cdn.example.com/seg1.ts", "https://cdn.example.com/seg2.ts"}
	data := map[string][]byte{
		"https://cdn.example.com/seg1.ts": []byte("d1"),
		"https://cdn.example.com/seg2.ts": []byte("d2"),
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "out.ts")

	// A sidecar from a totally different segment list.
	stale := &Manifest{Segments: []string{"https://other/x.ts"}, Completed: 1}
	if err := stale.Save(path + ".part"); err != nil {
		t.Fatal(err)
	}

	fetch := writeFetch(data)
	if err := FetchSegments(context.Background(), segments, FetchOptions{Concurrency: 2, Fetch: fetch}, path); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// File should contain both segments fresh — not stale data.
	got, _ := os.ReadFile(path)
	if string(got) != "d1d2" {
		t.Errorf("expected fresh 'd1d2', got %q", string(got))
	}
}

// newFlakyFetch returns a SegmentFetcher that fails the first N calls per URL,
// then returns the configured payload.
func newFlakyFetch(payloads map[string][]byte, failFirst map[string]int) SegmentFetcher {
	var mu sync.Mutex
	remaining := map[string]int{}
	for k, v := range failFirst {
		remaining[k] = v
	}
	return func(ctx context.Context, url string) ([]byte, error) {
		mu.Lock()
		left := remaining[url]
		if left > 0 {
			remaining[url]--
		}
		mu.Unlock()
		if left > 0 {
			return nil, errors.New("flaky transient failure")
		}
		return payloads[url], nil
	}
}
