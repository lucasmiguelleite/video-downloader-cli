// Package httpbody provides bounded HTTP response-body reads.
package httpbody

import (
	"fmt"
	"io"
)

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
		return nil, fmt.Errorf("response exceeds %d bytes", limit)
	}
	return data, nil
}
