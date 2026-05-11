package kick

import (
	"bufio"
	"net/url"
	"strings"
)

func IsMasterPlaylist(playlist string) bool {
	return strings.Contains(playlist, "#EXT-X-STREAM-INF")
}

type Variant struct {
	Resolution string
	URL        string
}

func ParseVariants(baseURL, playlist string) []Variant {
	var variants []Variant
	var currentRes string

	base, _ := url.Parse(baseURL)

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

		variantURL := resolveURL(base, line)
		variants = append(variants, Variant{
			Resolution: currentRes,
			URL:        variantURL,
		})
	}

	return variants
}

func SelectVariant(baseURL, playlist, quality string) string {
	variants := ParseVariants(baseURL, playlist)

	for _, v := range variants {
		if strings.Contains(v.Resolution, quality) || strings.HasSuffix(v.Resolution, "x"+strings.TrimSuffix(quality, "p")) {
			return v.URL
		}
	}

	if len(variants) > 0 {
		return variants[0].URL
	}

	return ""
}

func ParseSegments(baseURL, playlist string) []string {
	var segments []string

	base, _ := url.Parse(baseURL)

	scanner := bufio.NewScanner(strings.NewReader(playlist))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		segments = append(segments, resolveURL(base, line))
	}

	return segments
}

func extractResolution(tag string) string {
	for _, part := range strings.Split(tag, ",") {
		if strings.HasPrefix(part, "RESOLUTION=") {
			return strings.TrimPrefix(part, "RESOLUTION=")
		}
	}
	return ""
}

func resolveURL(base *url.URL, ref string) string {
	if strings.HasPrefix(ref, "http") {
		return ref
	}
	resolved, err := base.Parse(ref)
	if err != nil {
		return ref
	}
	return resolved.String()
}
