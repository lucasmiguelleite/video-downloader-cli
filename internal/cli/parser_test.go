package cli

import "testing"

func TestParseArgs_URLAndResolution(t *testing.T) {
	args := []string{"--url", "http://video", "--resolution", "1080p"}

	input, err := ParseArgs(args)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if input.URL != "http://video" {
		t.Errorf("expected URL 'http://video', got '%s'", input.URL)
	}

	if input.Quality != "1080p" {
		t.Errorf("expected resolution '1080p', got '%s'", input.Quality)
	}
}

func TestParseArgs_DefaultResolution(t *testing.T) {
	args := []string{"--url", "http://video"}

	input, err := ParseArgs(args)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if input.Quality != "720p" {
		t.Errorf("default resolution should be 720p, got '%s'", input.Quality)
	}
}

func TestParseArgs_Help(t *testing.T) {
	args := []string{"--help"}

	input, err := ParseArgs(args)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !input.ShowHelp {
		t.Error("ShowHelp should be true")
	}
}

func TestParseArgs_Version(t *testing.T) {
	args := []string{"--version"}

	input, err := ParseArgs(args)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !input.ShowVersion {
		t.Error("ShowVersion should be true")
	}
}

func TestParseArgs_EmptyURL(t *testing.T) {
	args := []string{}

	input, err := ParseArgs(args)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if input.URL != "" {
		t.Errorf("URL should be empty, got '%s'", input.URL)
	}
}

func TestParseArgs_OutputDir(t *testing.T) {
	args := []string{"--url", "http://video", "--output", "/tmp/videos"}

	input, err := ParseArgs(args)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if input.OutputDir != "/tmp/videos" {
		t.Errorf("expected OutputDir '/tmp/videos', got '%s'", input.OutputDir)
	}
}

func TestParseArgs_DefaultOutputDir(t *testing.T) {
	args := []string{"--url", "http://video"}

	input, err := ParseArgs(args)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if input.OutputDir != "" {
		t.Errorf("default OutputDir should be empty, got '%s'", input.OutputDir)
	}
}

func TestParseArgs_Concurrency(t *testing.T) {
	args := []string{"--url", "http://video", "--concurrency", "10"}

	input, err := ParseArgs(args)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if input.Concurrency != 10 {
		t.Errorf("expected Concurrency 10, got %d", input.Concurrency)
	}
}

func TestParseArgs_DefaultConcurrency(t *testing.T) {
	args := []string{"--url", "http://video"}

	input, err := ParseArgs(args)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if input.Concurrency != 20 {
		t.Errorf("default Concurrency should be 20, got %d", input.Concurrency)
	}
}

func TestParseArgs_InvalidFlag(t *testing.T) {
	args := []string{"--flag-inexistente"}

	_, err := ParseArgs(args)

	if err == nil {
		t.Error("expected error for invalid flag")
	}
}
