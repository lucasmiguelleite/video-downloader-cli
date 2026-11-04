package hls

import (
	"bufio"
	"fmt"
	"net/url"
	"strings"
)

// IsMasterPlaylist reports whether playlist declares stream variants.
func IsMasterPlaylist(playlist string) bool {
	return strings.Contains(playlist, "#EXT-X-STREAM-INF")
}

// Variant identifies a stream rendition in a master playlist.
type Variant struct {
	Resolution string
	URL        string
}

// ParseVariants parses the variants in a master playlist and resolves their
// URLs relative to baseURL.
func ParseVariants(baseURL, playlist string) ([]Variant, error) {
	var variants []Variant
	var currentRes string

	base, err := parseBaseURL(baseURL)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(strings.NewReader(playlist))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "#EXT-X-STREAM-INF") {
			currentRes = extractResolution(line)
			continue
		}

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		resolved, err := resolveURL(base, line)
		if err != nil {
			return nil, err
		}
		variants = append(variants, Variant{
			Resolution: currentRes,
			URL:        resolved,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan master playlist: %w", err)
	}
	return variants, nil
}

// SelectVariant returns the URL matching quality, or the first variant when
// an exact match is unavailable.
func SelectVariant(baseURL, playlist, quality string) (string, error) {
	variants, err := ParseVariants(baseURL, playlist)
	if err != nil {
		return "", err
	}

	for _, v := range variants {
		if strings.Contains(v.Resolution, quality) || strings.HasSuffix(v.Resolution, "x"+strings.TrimSuffix(quality, "p")) {
			return v.URL, nil
		}
	}

	if len(variants) > 0 {
		return variants[0].URL, nil
	}

	return "", nil
}

// ParseSegments parses a media playlist and resolves segment URLs relative to
// baseURL.
func ParseSegments(baseURL, playlist string) ([]string, error) {
	var segments []string

	base, err := parseBaseURL(baseURL)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(strings.NewReader(playlist))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		resolved, err := resolveURL(base, line)
		if err != nil {
			return nil, err
		}
		segments = append(segments, resolved)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan media playlist: %w", err)
	}
	return segments, nil
}

func extractResolution(tag string) string {
	for _, part := range strings.Split(tag, ",") {
		if strings.HasPrefix(part, "RESOLUTION=") {
			return strings.TrimPrefix(part, "RESOLUTION=")
		}
	}
	return ""
}

func parseBaseURL(rawURL string) (*url.URL, error) {
	base, err := url.Parse(rawURL)
	if err != nil || !base.IsAbs() || base.Host == "" {
		return nil, fmt.Errorf("invalid playlist base URL %q", rawURL)
	}
	return base, nil
}

func resolveURL(base *url.URL, ref string) (string, error) {
	resolved, err := base.Parse(ref)
	if err != nil {
		return "", fmt.Errorf("resolve playlist URL %q: %w", ref, err)
	}
	if !resolved.IsAbs() || resolved.Host == "" {
		return "", fmt.Errorf("resolved playlist URL %q is not absolute", ref)
	}
	return resolved.String(), nil
}
