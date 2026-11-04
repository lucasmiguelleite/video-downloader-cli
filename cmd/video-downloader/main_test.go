package main

import (
	"context"
	"testing"

	"youtube-downloader/internal/downloader"
)

type factoryService struct{ provider string }

func (s *factoryService) GetVideo(context.Context, string) (*downloader.Video, error) {
	return &downloader.Video{}, nil
}

func (s *factoryService) Download(context.Context, *downloader.Video, string, string) error {
	return nil
}

func TestProviderForURL_RejectsLookalikeHost(t *testing.T) {
	_, err := providerForURL("https://notkick.com/channel/videos/123")
	if err == nil {
		t.Fatal("expected lookalike host to be rejected")
	}
}

func TestProviderForURL_RejectsHTTP(t *testing.T) {
	_, err := providerForURL("http://www.youtube.com/watch?v=x")
	if err == nil {
		t.Fatal("expected HTTP URL to be rejected")
	}
}

func TestProviderForURL_RecognizesSupportedHosts(t *testing.T) {
	tests := map[string]string{
		"https://www.youtube.com/watch?v=x": "youtube",
		"https://player.twitch.tv/?video=1": "twitch",
		"https://kick.com/a/videos/id":      "kick",
	}
	for rawURL, want := range tests {
		got, err := providerForURL(rawURL)
		if err != nil {
			t.Errorf("providerForURL(%q): %v", rawURL, err)
			continue
		}
		if got != want {
			t.Errorf("providerForURL(%q) = %q, want %q", rawURL, got, want)
		}
	}
}

func TestResolveServiceWithFactories_SelectsFactory(t *testing.T) {
	factories := serviceFactories{
		kick:    func(int) downloader.Service { return &factoryService{provider: "kick"} },
		twitch:  func(int) downloader.Service { return &factoryService{provider: "twitch"} },
		youtube: func(int) downloader.Service { return &factoryService{provider: "youtube"} },
	}
	tests := []struct {
		url        string
		provider   string
		needsRemux bool
	}{
		{"https://kick.com/a/videos/id", "kick", true},
		{"https://twitch.tv/videos/1", "twitch", true},
		{"https://youtube.com/watch?v=x", "youtube", false},
	}
	for _, test := range tests {
		service, needsRemux, err := resolveServiceWithFactories(test.url, 7, factories)
		if err != nil {
			t.Errorf("resolveServiceWithFactories(%q): %v", test.url, err)
			continue
		}
		if service.(*factoryService).provider != test.provider || needsRemux != test.needsRemux {
			t.Errorf("url %q selected %#v, needsRemux=%v", test.url, service, needsRemux)
		}
	}
}

func TestResolveServiceWithFactories_InvalidURLCallsNoFactory(t *testing.T) {
	called := false
	factory := func(int) downloader.Service { called = true; return &factoryService{} }
	_, _, err := resolveServiceWithFactories("https://example.com/video", 7, serviceFactories{
		kick: factory, twitch: factory, youtube: factory,
	})
	if err == nil {
		t.Fatal("expected unsupported host error")
	}
	if called {
		t.Fatal("factory should not be called for invalid URL")
	}
}
