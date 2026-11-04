package kick

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestRealAPI_FetchPlaylist(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "#EXTM3U")
	}))
	t.Cleanup(server.Close)

	playlist, err := NewRealAPIWithHTTPClient(server.Client()).FetchPlaylist(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("FetchPlaylist: %v", err)
	}
	if playlist != "#EXTM3U" {
		t.Errorf("playlist = %q", playlist)
	}
}

func TestRealAPI_FetchPlaylistRejectsNonSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(server.Close)

	_, err := NewRealAPIWithHTTPClient(server.Client()).FetchPlaylist(context.Background(), server.URL)
	if err == nil {
		t.Fatal("expected HTTP status error")
	}
}

func TestRealAPI_FetchPlaylistHonorsCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	t.Cleanup(server.Close)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := NewRealAPIWithHTTPClient(server.Client()).FetchPlaylist(ctx, server.URL)
	if err == nil {
		t.Fatal("expected cancellation error")
	}
}

func TestRealAPI_FetchVideoInfoRejectsMalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "{")
	}))
	t.Cleanup(server.Close)
	client := server.Client()
	client.Transport = rewriteTransport(t, server.URL, client.Transport)
	_, err := NewRealAPIWithHTTPClient(client).FetchVideoInfo(context.Background(), "video-id")
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
