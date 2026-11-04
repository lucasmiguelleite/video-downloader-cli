package downloader

import (
	"context"
	"math/rand/v2"
	"time"
)

// RetryConfig controls WithRetry timing. Zero Base and Max values use the
// production defaults. A zero Jitter explicitly disables jitter.
type RetryConfig struct {
	// Base is the first backoff sleep. Defaults to 500ms.
	Base time.Duration
	// Max caps a single backoff sleep. Defaults to 30s.
	Max time.Duration
	// Jitter is the +/- random spread added to each sleep. Zero disables it.
	Jitter time.Duration
}

var defaultRetryConfig = RetryConfig{
	Base:   500 * time.Millisecond,
	Max:    30 * time.Second,
	Jitter: 250 * time.Millisecond,
}

func (c RetryConfig) withDefaults() RetryConfig {
	cfg := c
	if cfg.Base <= 0 {
		cfg.Base = defaultRetryConfig.Base
	}
	if cfg.Max <= 0 {
		cfg.Max = defaultRetryConfig.Max
	}
	return cfg
}

// WithRetry calls fn up to attempts times, retrying while fn returns a non-nil
// error. Between attempts it sleeps with exponential backoff (Base*2^k, capped
// at Max) plus a small random jitter. The backoff sleep is interruptible: if
// ctx is cancelled, WithRetry returns ctx.Err() immediately.
//
// attempts must be >= 1. If fn returns nil, WithRetry returns nil right away
// (no extra calls).
//
// Uses the package's immutable production defaults. Use WithRetryConfig to
// customise timing.
func WithRetry(ctx context.Context, attempts int, fn func() error) error {
	return WithRetryConfig(ctx, attempts, defaultRetryConfig, fn)
}

// WithRetryConfig is the configurable variant. Zero Base and Max fields use
// production defaults; zero Jitter disables random jitter.
func WithRetryConfig(ctx context.Context, attempts int, cfg RetryConfig, fn func() error) error {
	if attempts < 1 {
		attempts = 1
	}
	cfg = cfg.withDefaults()

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		// Honour a cancelled context before doing work.
		if err := ctx.Err(); err != nil {
			return err
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		// No need to sleep after the final attempt.
		if attempt == attempts {
			break
		}

		backoff := backoffFor(attempt, cfg)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}
	}
	return lastErr
}

// backoffFor computes the sleep duration for the given (1-based) attempt
// number: base*2^(attempt-1), capped at max, with +/- jitter.
func backoffFor(attempt int, cfg RetryConfig) time.Duration {
	d := cfg.Base
	for i := 1; i < attempt; i++ {
		d *= 2
		if d > cfg.Max {
			d = cfg.Max
			break
		}
	}
	if cfg.Jitter > 0 {
		// rand.Int64N is safe across goroutines in math/rand/v2.
		jitter := time.Duration(rand.Int64N(int64(2*cfg.Jitter))) - cfg.Jitter
		d += jitter
	}
	if d < 0 {
		d = 0
	}
	return d
}
