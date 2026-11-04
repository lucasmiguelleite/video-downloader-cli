package main

import "testing"

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
