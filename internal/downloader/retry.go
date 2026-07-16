package downloader

import (
	"context"
	"math/rand/v2"
	"time"
)

// RetryConfig controls WithRetry timing. The zero value falls back to sane
// defaults at call time, so callers only need to set the fields they care
// about.
type RetryConfig struct {
	// Base is the first backoff sleep. Defaults to 500ms.
	Base time.Duration
	// Max caps a single backoff sleep. Defaults to 30s.
	Max time.Duration
	// Jitter is the +/- random spread added to each sleep. Defaults to 250ms.
	Jitter time.Duration
}

// DefaultRetryConfig is the backoff configuration used by WithRetry when no
// override is supplied. Tests may temporarily shorten it to keep the suite
// fast; production code should leave it alone.
var DefaultRetryConfig = RetryConfig{
	Base:   500 * time.Millisecond,
	Max:    30 * time.Second,
	Jitter: 250 * time.Millisecond,
}

func (c RetryConfig) withDefaults() RetryConfig {
	cfg := c
	if cfg.Base <= 0 {
		cfg.Base = DefaultRetryConfig.Base
	}
	if cfg.Max <= 0 {
		cfg.Max = DefaultRetryConfig.Max
	}
	// Jitter may legitimately be 0 (disable), so only fall back when unset
	// AND identical to the zero value would matter — keep it explicit:
	if cfg.Jitter < 0 {
		cfg.Jitter = DefaultRetryConfig.Jitter
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
// Uses DefaultRetryConfig. Use WithRetryConfig to customise timing.
func WithRetry(ctx context.Context, attempts int, fn func() error) error {
	return WithRetryConfig(ctx, attempts, DefaultRetryConfig, fn)
}

// WithRetryConfig is the configurable variant. cfg's zero fields fall back to
// the defaults from DefaultRetryConfig.
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
