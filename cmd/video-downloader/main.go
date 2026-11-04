package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"youtube-downloader/internal/app"
	"youtube-downloader/internal/cli"
	"youtube-downloader/internal/downloader"
	"youtube-downloader/internal/kick"
	"youtube-downloader/internal/twitch"
	"youtube-downloader/internal/youtube"
)

func main() {
	args := os.Args[1:]

	input, err := cli.ParseArgs(args)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	if input.ShowHelp {
		fmt.Printf("video-downloader v%s\n\n", cli.Version)
		fmt.Println("Download videos from YouTube, Twitch, and Kick from the command line.")
		fmt.Println()
		fmt.Println("Usage: video-downloader --url <URL> [--resolution <resolution>]")
		fmt.Println()
		fmt.Println("Flags:")
		fmt.Println("  --url <URL>         Video URL (required) — YouTube, Twitch, Kick")
		fmt.Println("  --resolution <res>  Video resolution (default: 720p)")
		fmt.Println("  --output <dir>      Directory to save the video (default: ~/Downloads)")
		fmt.Println("  --concurrency <N>   Parallel segment downloads for HLS streams (default: 20)")
		fmt.Println("  --verbose           Enable debug-level logging")
		fmt.Println("  --help              Show this help")
		fmt.Println("  --version           Show version")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  video-downloader --url 'https://youtube.com/watch?v=xxx'")
		fmt.Println("  video-downloader --url 'https://twitch.tv/videos/1234567890'")
		fmt.Println("  video-downloader --url 'https://kick.com/nitrao/videos/7b0877f1-983d-4c5f-932e-cda88e8196f1'")
		fmt.Println("  video-downloader --url 'https://youtube.com/watch?v=xxx' --resolution 1080p")
		fmt.Println("  video-downloader --url 'https://twitch.tv/videos/1234567890' --output ~/Videos")
		os.Exit(0)
	}

	if input.ShowVersion {
		fmt.Printf("video-downloader v%s\n", cli.Version)
		os.Exit(0)
	}

	if input.URL == "" {
		fmt.Println("Error: --url flag is required. Use --help for more information.")
		os.Exit(1)
	}

	runningInBackground, err := cli.IsRunningInBackground(os.Stdin)
	if err != nil {
		fmt.Println("Error checking terminal:", err)
		os.Exit(1)
	}
	if runningInBackground {
		fmt.Fprintln(os.Stderr, "Warning: download started in background.")
		fmt.Fprintln(os.Stderr, "If the URL contains '&', wrap it in quotes:")
		fmt.Fprintln(os.Stderr, `  video-downloader --url 'https://www.youtube.com/watch?v=gkx519Vi-co&t=2824s'`)
		os.Exit(1)
	}

	service, needsRemux, err := resolveService(input.URL, input.Concurrency)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	logger := cli.NewLogger(os.Stderr, input.Verbose)
	slog.SetDefault(logger)

	var usecase *app.DownloadUseCase
	if needsRemux && input.OutputDir != "" {
		usecase = app.NewDownloadUseCaseWithRemux(service, input.OutputDir)
	} else if needsRemux {
		usecase = app.NewDownloadUseCaseWithRemux(service, "")
	} else if input.OutputDir != "" {
		usecase = app.NewDownloadUseCaseWithDir(service, input.OutputDir)
	} else {
		usecase = app.NewDownloadUseCase(service)
	}
	usecase.WithLogger(logger)

	// Catch Ctrl+C / SIGTERM so in-flight HTTP requests cancel cleanly and
	// the partial file is kept for resume.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	path, err := usecase.Execute(ctx, input.URL, input.Quality)
	if err != nil {
		fmt.Println("Error downloading video:", err)
		os.Exit(1)
	}

	fmt.Println("Download completed successfully!")
	fmt.Printf("Saved to: %s\n", path)
}

func resolveService(rawURL string, concurrency int) (downloader.Service, bool, error) {
	provider, err := providerForURL(rawURL)
	if err != nil {
		return nil, false, err
	}
	switch provider {
	case "kick":
		return kick.NewClientWithConcurrency(kick.NewRealAPI(), concurrency), true, nil
	case "twitch":
		return twitch.NewClientWithConcurrency(twitch.NewRealAPI(), concurrency), true, nil
	case "youtube":
		return youtube.NewClient(youtube.NewRealAPI()), false, nil
	default:
		return nil, false, fmt.Errorf("unsupported video provider %q", provider)
	}
}

func providerForURL(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" {
		return "", fmt.Errorf("invalid video URL %q; expected an absolute HTTPS URL", rawURL)
	}
	host := parsed.Hostname()
	switch host {
	case "kick.com", "www.kick.com":
		return "kick", nil
	case "twitch.tv", "www.twitch.tv", "player.twitch.tv":
		return "twitch", nil
	case "youtube.com", "www.youtube.com", "m.youtube.com", "youtu.be", "www.youtu.be":
		return "youtube", nil
	default:
		return "", fmt.Errorf("unsupported video host %q", host)
	}
}
