package cli

import (
	"flag"
	"fmt"
	"os"
)

const Version = "0.1.1"

type Input struct {
	URL         string
	Quality     string
	OutputDir   string
	Concurrency int
	ShowHelp    bool
	ShowVersion bool
}

func ParseArgs(args []string) (Input, error) {
	fs := flag.NewFlagSet("youtube-downloader", flag.ContinueOnError)

	url := fs.String("url", "", "Video URL")
	quality := fs.String("resolution", "720p", "Video resolution (e.g. 360p, 480p, 720p, 1080p)")
	outputDir := fs.String("output", "", "Directory to save the video (default: ~/Downloads)")
	concurrency := fs.Int("concurrency", 20, "Number of parallel segment downloads (Kick only)")
	showHelp := fs.Bool("help", false, "Show help")
	showVersion := fs.Bool("version", false, "Show version")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: youtube-downloader --url <URL> [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  youtube-downloader --url 'https://youtube.com/watch?v=xxx'\n")
		fmt.Fprintf(os.Stderr, "  youtube-downloader --url 'https://youtube.com/watch?v=xxx' --resolution 1080p\n")
	}

	if err := fs.Parse(args); err != nil {
		return Input{}, err
	}

	input := Input{
		URL:         *url,
		Quality:     *quality,
		OutputDir:   *outputDir,
		Concurrency: *concurrency,
		ShowHelp:    *showHelp,
		ShowVersion: *showVersion,
	}

	return input, nil
}
