package youtube

import (
	"net/http"
	"testing"
)

func TestNewRealAPIWithHTTPClient_UsesInjectedClient(t *testing.T) {
	client := &http.Client{}
	api := NewRealAPIWithHTTPClient(client)
	if api.client.HTTPClient != client {
		t.Fatal("injected HTTP client was not used")
	}
}

func TestNewRealAPIWithHTTPClient_NilUsesDefaultClient(t *testing.T) {
	api := NewRealAPIWithHTTPClient(nil)
	if api.client.HTTPClient == nil {
		t.Fatal("expected a default HTTP client")
	}
}
