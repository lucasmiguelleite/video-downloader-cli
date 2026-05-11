package cli

import (
	"flag"
	"fmt"
	"os"
)

const Version = "0.0.1"

type Input struct {
	URL         string
	Quality     string
	ShowHelp    bool
	ShowVersion bool
}

func ParseArgs(args []string) (Input, error) {
	fs := flag.NewFlagSet("youtube-downloader", flag.ContinueOnError)

	url := fs.String("url", "", "URL do vídeo do YouTube")
	quality := fs.String("resolution", "720p", "Resolução do vídeo (ex: 360p, 480p, 720p, 1080p)")
	showHelp := fs.Bool("help", false, "Exibe a ajuda")
	showVersion := fs.Bool("version", false, "Exibe a versão")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Uso: youtube-downloader --url <URL> [opções]\n\n")
		fmt.Fprintf(os.Stderr, "Opções:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExemplos:\n")
		fmt.Fprintf(os.Stderr, "  youtube-downloader --url 'https://youtube.com/watch?v=xxx'\n")
		fmt.Fprintf(os.Stderr, "  youtube-downloader --url 'https://youtube.com/watch?v=xxx' --resolution 1080p\n")
	}

	if err := fs.Parse(args); err != nil {
		return Input{}, err
	}

	input := Input{
		URL:         *url,
		Quality:     *quality,
		ShowHelp:    *showHelp,
		ShowVersion: *showVersion,
	}

	return input, nil
}
