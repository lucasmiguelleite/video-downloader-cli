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
