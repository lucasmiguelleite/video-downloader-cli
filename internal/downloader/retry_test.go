package downloader

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestWithRetry_SucceedsOnFirstAttempt(t *testing.T) {
	var calls int32
	err := WithRetry(context.Background(), 3, func() error {
		atomic.AddInt32(&calls, 1)
		return nil
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

// fastCfg keeps the slow-backoff tests fast and deterministic.
var fastCfg = RetryConfig{Base: time.Millisecond, Max: 5 * time.Millisecond, Jitter: 0}

func TestWithRetry_SucceedsOnLaterAttempt(t *testing.T) {
	var calls int32
	errSentinel := errors.New("transient")
	err := WithRetryConfig(context.Background(), 3, fastCfg, func() error {
		n := atomic.AddInt32(&calls, 1)
		if n < 3 {
			return errSentinel
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestWithRetry_ExhaustsAttemptsAndReturnsLastError(t *testing.T) {
	var calls int32
	errSentinel := errors.New("permanent")
	err := WithRetryConfig(context.Background(), 3, fastCfg, func() error {
		atomic.AddInt32(&calls, 1)
		return errSentinel
	})
	if !errors.Is(err, errSentinel) {
		t.Errorf("expected wrapped sentinel error, got: %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestWithRetry_NilFnReturnsNil(t *testing.T) {
	var calls int32
	// Returning nil means "no error" — the loop should stop after 1 call.
	err := WithRetry(context.Background(), 5, func() error {
		atomic.AddInt32(&calls, 1)
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestWithRetry_RespectsCancelledContextDuringBackoff(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before first attempt

	var calls int32
	errSentinel := errors.New("fail")
	err := WithRetry(ctx, 5, func() error {
		atomic.AddInt32(&calls, 1)
		return errSentinel
	})

	// After the first failing attempt, the backoff sleep should observe the
	// cancelled context and abort, returning the context error.
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got: %v", err)
	}
}

func TestWithRetry_BackoffIsCancellable(t *testing.T) {
	// If we cancel mid-backoff, the function should return well before the
	// full backoff would have slept. This guards against a naive time.Sleep.
	ctx, cancel := context.WithCancel(context.Background())
	errSentinel := errors.New("fail")

	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	err := WithRetry(ctx, 10, func() error { return errSentinel })
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected error from exhaustion/cancellation")
	}
	// Default exponential backoff for 10 attempts would take minutes; if the
	// backoff honours ctx, we should return within a couple of seconds.
	if elapsed > 2*time.Second {
		t.Errorf("backoff was not cancellable: took %v", elapsed)
	}
}

func TestDefaultRetryConfig_HasSaneProductionValues(t *testing.T) {
	// Guard against accidentally zeroing the production backoff.
	if defaultRetryConfig.Base <= 0 {
		t.Errorf("expected positive base backoff, got %v", defaultRetryConfig.Base)
	}
	if defaultRetryConfig.Max < defaultRetryConfig.Base {
		t.Errorf("max backoff %v should be >= base %v", defaultRetryConfig.Max, defaultRetryConfig.Base)
	}
}

func TestRetryConfig_ZeroFieldsFallBackToDefaults(t *testing.T) {
	cfg := RetryConfig{}.withDefaults()
	if cfg.Base != defaultRetryConfig.Base {
		t.Errorf("expected default base, got %v", cfg.Base)
	}
	if cfg.Max != defaultRetryConfig.Max {
		t.Errorf("expected default max, got %v", cfg.Max)
	}
	// Explicit zero jitter should be preserved (it's a valid "disable jitter".
	cfg = RetryConfig{Base: time.Second, Max: time.Second, Jitter: 0}.withDefaults()
	if cfg.Jitter != 0 {
		t.Errorf("explicit Jitter=0 should be preserved, got %v", cfg.Jitter)
	}
}
