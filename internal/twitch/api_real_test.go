package twitch

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"youtube-downloader/internal/httpbody"
)

func TestRealAPI_FetchSegment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "segment")
	}))
	t.Cleanup(server.Close)

	segment, err := NewRealAPIWithHTTPClient(server.Client()).FetchSegment(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("FetchSegment: %v", err)
	}
	t.Cleanup(func() { _ = segment.Close() })
	data, err := io.ReadAll(segment)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "segment" {
		t.Errorf("segment = %q", data)
	}
}

func TestRealAPI_FetchSegmentRejectsOversizedStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.CopyN(w, zeroReader{}, maxResponseBytes+1)
	}))
	t.Cleanup(server.Close)
	body, err := NewRealAPIWithHTTPClient(server.Client()).FetchSegment(context.Background(), server.URL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = body.Close() })
	_, err = io.Copy(io.Discard, body)
	if !errors.Is(err, httpbody.ErrResponseTooLarge) {
		t.Fatalf("expected ErrResponseTooLarge, got %v", err)
	}
}

type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}

func TestRealAPI_FetchSegmentRejectsRedirectStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Location", "/same")
		w.WriteHeader(http.StatusTemporaryRedirect)
	}))
	t.Cleanup(server.Close)
	client := server.Client()
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }

	_, err := NewRealAPIWithHTTPClient(client).FetchSegment(context.Background(), server.URL)
	if err == nil {
		t.Fatal("expected redirect status error")
	}
}

func TestRealAPI_FetchVideoInfoRejectsMalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "{")
	}))
	t.Cleanup(server.Close)
	client := server.Client()
	client.Transport = rewriteTransport(t, server.URL, client.Transport)
	_, err := NewRealAPIWithHTTPClient(client).FetchVideoInfo(context.Background(), "123")
	if err == nil {
		t.Fatal("expected malformed JSON error")
	}
}

func rewriteTransport(t *testing.T, rawURL string, base http.RoundTripper) http.RoundTripper {
	t.Helper()
	target, err := url.Parse(rawURL)
	if err != nil {
		t.Fatal(err)
	}
	return roundTripFunc(func(req *http.Request) (*http.Response, error) {
		clone := req.Clone(req.Context())
		clone.URL.Scheme = target.Scheme
		clone.URL.Host = target.Host
		return base.RoundTrip(clone)
	})
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return fn(req) }
