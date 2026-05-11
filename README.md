# youtube-downloader

CLI em Go para baixar vídeos do YouTube via linha de comando.

## Instalação

### Via `go install`

```bash
go install ./cmd/youtube-downloader/
```

O binário fica disponível em `$GOPATH/bin` (ou `$HOME/go/bin`).

### Via `go build`

```bash
go build -o bin/youtube-downloader ./cmd/youtube-downloader/
```

## Uso

```bash
# Baixar com resolução padrão (720p)
youtube-downloader --url 'https://www.youtube.com/watch?v=dQw4w9WgXcQ'

# Especificar resolução
youtube-downloader --url 'https://www.youtube.com/watch?v=dQw4w9WgXcQ' --resolution 1080p

# Ver ajuda
youtube-downloader --help

# Ver versão
youtube-downloader --version
```

### Flags

| Flag | Descrição | Padrão |
|---|---|---|
| `--url` | URL do vídeo do YouTube (obrigatório) | — |
| `--resolution` | Resolução do vídeo | `720p` |
| `--help` | Exibe a ajuda | — |
| `--version` | Exibe a versão | — |

> Se a URL contiver `&`, coloque-a entre aspas simples para evitar problemas com o shell.

## Desenvolvimento

### Pré-requisitos

- Go 1.26+

### Rodando

```bash
go run ./cmd/youtube-downloader/ --url 'https://www.youtube.com/watch?v=dQw4w9WgXcQ'
```

### Testes

```bash
go test ./...
```

### Build

```bash
go build -o bin/youtube-downloader ./cmd/youtube-downloader/
```

## Estrutura do projeto

```
cmd/youtube-downloader/    # Entry point (main)
internal/
  app/                     # Casos de uso (regra de negócio)
  cli/                     # Parser de argumentos e utilitários de terminal
  downloader/              # Interface do serviço de download
  fs/                      # Salvar arquivos em disco
  youtube/                 # Cliente da API do YouTube
```

O projeto segue os princípios de Clean Architecture e SOLID, com inversão de dependência entre as camadas.
