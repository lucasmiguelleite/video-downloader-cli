// Package httpbody provides bounded HTTP response-body reads.
package httpbody

import (
	"errors"
	"fmt"
	"io"
)

// ErrResponseTooLarge indicates an HTTP body exceeded its configured limit.
var ErrResponseTooLarge = errors.New("response exceeds configured limit")

// Read returns at most limit bytes and errors when body exceeds that limit.
func Read(body io.Reader, limit int64) ([]byte, error) {
	if limit < 1 {
		return nil, fmt.Errorf("response limit must be positive, got %d", limit)
	}
	data, err := io.ReadAll(io.LimitReader(body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("%w of %d bytes", ErrResponseTooLarge, limit)
	}
	return data, nil
}

// Limit returns a streaming body that errors after limit bytes.
func Limit(body io.ReadCloser, limit int64) (io.ReadCloser, error) {
	if limit < 1 {
		return nil, fmt.Errorf("response limit must be positive, got %d", limit)
	}
	return &limitedReadCloser{body: body, remaining: limit, limit: limit}, nil
}

type limitedReadCloser struct {
	body      io.ReadCloser
	remaining int64
	limit     int64
}

func (r *limitedReadCloser) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	maxRead := int64(len(p))
	if maxRead > r.remaining+1 {
		maxRead = r.remaining + 1
	}
	n, err := r.body.Read(p[:maxRead])
	if int64(n) > r.remaining {
		allowed := int(r.remaining)
		r.remaining = 0
		return allowed, fmt.Errorf("%w of %d bytes", ErrResponseTooLarge, r.limit)
	}
	r.remaining -= int64(n)
	return n, err
}

func (r *limitedReadCloser) Close() error { return r.body.Close() }
