package hls

import "testing"

func TestParseSegments(t *testing.T) {
	playlist := `#EXTM3U
#EXT-X-TARGETDURATION:10
#EXTINF:10.0,
https://cdn.example.com/seg1.ts
#EXTINF:10.0,
https://cdn.example.com/seg2.ts
#EXTINF:5.0,
https://cdn.example.com/seg3.ts
#EXT-X-ENDLIST`

	segments := ParseSegments("https://cdn.example.com/playlist.m3u8", playlist)

	if len(segments) != 3 {
		t.Fatalf("expected 3 segments, got %d", len(segments))
	}

	expected := []string{
		"https://cdn.example.com/seg1.ts",
		"https://cdn.example.com/seg2.ts",
		"https://cdn.example.com/seg3.ts",
	}

	for i, seg := range segments {
		if seg != expected[i] {
			t.Errorf("segment %d: expected %s, got %s", i, expected[i], seg)
		}
	}
}

func TestParseSegments_RelativeURLs(t *testing.T) {
	playlist := `#EXTM3U
#EXT-X-TARGETDURATION:10
#EXTINF:10.0,
segment1.ts
#EXTINF:10.0,
segment2.ts
#EXT-X-ENDLIST`

	segments := ParseSegments("https://cdn.example.com/video/playlist.m3u8", playlist)

	if len(segments) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(segments))
	}

	if segments[0] != "https://cdn.example.com/video/segment1.ts" {
		t.Errorf("expected absolute URL, got %s", segments[0])
	}

	if segments[1] != "https://cdn.example.com/video/segment2.ts" {
		t.Errorf("expected absolute URL, got %s", segments[1])
	}
}

func TestParseSegments_EmptyPlaylist(t *testing.T) {
	segments := ParseSegments("https://example.com/playlist.m3u8", "")

	if len(segments) != 0 {
		t.Errorf("expected 0 segments, got %d", len(segments))
	}
}

func TestIsMasterPlaylist(t *testing.T) {
	master := `#EXTM3U
#EXT-X-STREAM-INF:BANDWIDTH=4000000,RESOLUTION=1920x1080
https://cdn.example.com/1080p/index.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=2000000,RESOLUTION=1280x720
https://cdn.example.com/720p/index.m3u8`

	if !IsMasterPlaylist(master) {
		t.Error("expected master playlist to be detected")
	}

	media := `#EXTM3U
#EXT-X-TARGETDURATION:10
#EXTINF:10.0,
segment1.ts`

	if IsMasterPlaylist(media) {
		t.Error("media playlist should not be detected as master")
	}
}

func TestSelectVariant(t *testing.T) {
	master := `#EXTM3U
#EXT-X-STREAM-INF:BANDWIDTH=4000000,RESOLUTION=1920x1080
https://cdn.example.com/1080p/index.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=2000000,RESOLUTION=1280x720
https://cdn.example.com/720p/index.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=1000000,RESOLUTION=854x480
https://cdn.example.com/480p/index.m3u8`

	url := SelectVariant("https://cdn.example.com/master.m3u8", master, "720p")

	if url != "https://cdn.example.com/720p/index.m3u8" {
		t.Errorf("expected 720p variant, got %s", url)
	}
}

func TestSelectVariant_DefaultToFirst(t *testing.T) {
	master := `#EXTM3U
#EXT-X-STREAM-INF:BANDWIDTH=4000000,RESOLUTION=1920x1080
https://cdn.example.com/1080p/index.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=2000000,RESOLUTION=1280x720
https://cdn.example.com/720p/index.m3u8`

	url := SelectVariant("https://cdn.example.com/master.m3u8", master, "4k")

	if url != "https://cdn.example.com/1080p/index.m3u8" {
		t.Errorf("expected first variant as fallback, got %s", url)
	}
}
