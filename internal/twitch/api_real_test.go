package twitch

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
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
	if string(segment) != "segment" {
		t.Errorf("segment = %q", segment)
	}
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
