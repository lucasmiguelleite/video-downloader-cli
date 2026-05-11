package main

import (
	"fmt"
	"os"

	"youtube-downloader/internal/app"
	"youtube-downloader/internal/cli"
	"youtube-downloader/internal/youtube"
)

func main() {
	args := os.Args[1:]

	input, err := cli.ParseArgs(args)
	if err != nil {
		fmt.Println("Erro:", err)
		os.Exit(1)
	}

	if input.ShowHelp {
		fmt.Printf("youtube-downloader v%s\n\n", cli.Version)
		fmt.Println("Baixe vídeos do YouTube via linha de comando.")
		fmt.Println()
		fmt.Println("Uso: youtube-downloader --url <URL> [--resolution <resolução>]")
		fmt.Println()
		fmt.Println("Flags:")
		fmt.Println("  --url <URL>         URL do vídeo do YouTube (obrigatório)")
		fmt.Println("  --resolution <res>  Resolução do vídeo (padrão: 720p)")
		fmt.Println("  --help              Exibe esta ajuda")
		fmt.Println("  --version           Exibe a versão")
		fmt.Println()
		fmt.Println("Exemplos:")
		fmt.Println("  youtube-downloader --url 'https://youtube.com/watch?v=xxx'")
		fmt.Println("  youtube-downloader --url 'https://youtube.com/watch?v=xxx' --resolution 1080p")
		os.Exit(0)
	}

	if input.ShowVersion {
		fmt.Printf("youtube-downloader v%s\n", cli.Version)
		os.Exit(0)
	}

	if input.URL == "" {
		fmt.Println("Erro: flag --url é obrigatória. Use --help para mais informações.")
		os.Exit(1)
	}

	runningInBackground, err := cli.IsRunningInBackground(os.Stdin)
	if err != nil {
		fmt.Println("Erro ao verificar terminal:", err)
		os.Exit(1)
	}
	if runningInBackground {
		fmt.Fprintln(os.Stderr, "Aviso: o download foi iniciado em segundo plano.")
		fmt.Fprintln(os.Stderr, "Se a URL tiver '&', coloque a URL inteira entre aspas:")
		fmt.Fprintln(os.Stderr, `  youtube-downloader --url 'https://www.youtube.com/watch?v=gkx519Vi-co&t=2824s'`)
		os.Exit(1)
	}

	api := youtube.NewRealAPI()
	ytClient := youtube.NewClient(api)
	usecase := app.NewDownloadUseCase(ytClient)

	err = usecase.Execute(input.URL, input.Quality)
	if err != nil {
		fmt.Println("Erro ao baixar vídeo:", err)
		os.Exit(1)
	}

	fmt.Println("Download concluído com sucesso!")
}
