package main

import (
	"fmt"
	"os"
	"strings"

	"youtube-downloader/internal/app"
	"youtube-downloader/internal/cli"
	"youtube-downloader/internal/downloader"
	"youtube-downloader/internal/kick"
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
		fmt.Println("Download videos from YouTube and Kick from the command line.")
		fmt.Println()
		fmt.Println("Usage: video-downloader --url <URL> [--resolution <resolution>]")
		fmt.Println()
		fmt.Println("Flags:")
		fmt.Println("  --url <URL>         Video URL (required) — supports YouTube and Kick")
		fmt.Println("  --resolution <res>  Video resolution (default: 720p)")
		fmt.Println("  --output <dir>      Directory to save the video (default: ~/Downloads)")
		fmt.Println("  --help              Show this help")
		fmt.Println("  --version           Show version")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  video-downloader --url 'https://youtube.com/watch?v=xxx'")
		fmt.Println("  video-downloader --url 'https://kick.com/nitrao/videos/7b0877f1-983d-4c5f-932e-cda88e8196f1'")
		fmt.Println("  video-downloader --url 'https://youtube.com/watch?v=xxx' --resolution 1080p")
		fmt.Println("  video-downloader --url 'https://youtube.com/watch?v=xxx' --output ~/Videos")
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

	service := resolveService(input.URL)

	var usecase *app.DownloadUseCase
	if input.OutputDir != "" {
		usecase = app.NewDownloadUseCaseWithDir(service, input.OutputDir)
	} else {
		usecase = app.NewDownloadUseCase(service)
	}

	path, err := usecase.Execute(input.URL, input.Quality)
	if err != nil {
		fmt.Println("Error downloading video:", err)
		os.Exit(1)
	}

	fmt.Println("Download completed successfully!")
	fmt.Printf("Saved to: %s\n", path)
}

func resolveService(url string) downloader.Service {
	if strings.Contains(url, "kick.com") {
		return kick.NewClient(kick.NewRealAPI())
	}
	return youtube.NewClient(youtube.NewRealAPI())
}
