package main

import (
	"fmt"
	"os"

	"youtube-downloader/internal/app"
	"youtube-downloader/internal/cli"
	"youtube-downloader/internal/youtube"
)

func main() {
	// 1. Pegar os argumentos da CLI
	args := os.Args[1:]

	input, err := cli.ParseArgs(args)
	if err != nil {
		fmt.Println("Erro:", err)
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
		fmt.Fprintln(os.Stderr, `go run ./cmd/youtube-downloader/main.go 'https://www.youtube.com/watch?v=gkx519Vi-co&t=2824s'`)
		os.Exit(1)
	}

	// 2. Inicializar dependências
	api := youtube.NewRealAPI()
	ytClient := youtube.NewClient(api)

	// 3. Injetar no use case
	usecase := app.NewDownloadUseCase(ytClient)

	// 4. Executar fluxo principal
	err = usecase.Execute(input.URL, input.Quality)
	if err != nil {
		fmt.Println("Erro ao baixar vídeo:", err)
		os.Exit(1)
	}

	fmt.Println("Download concluído com sucesso!")
}
