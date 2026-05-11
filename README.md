# youtube-downloader

A Go CLI to download YouTube videos from the command line.

## Installation

### Via `go install`

```bash
go install ./cmd/youtube-downloader/
```

The binary becomes available at `$GOPATH/bin` (or `$HOME/go/bin`).

### Via `go build`

```bash
go build -o bin/youtube-downloader ./cmd/youtube-downloader/
```

## Usage

```bash
# Download with default resolution (720p)
youtube-downloader --url 'https://www.youtube.com/watch?v=dQw4w9WgXcQ'

# Specify resolution
youtube-downloader --url 'https://www.youtube.com/watch?v=dQw4w9WgXcQ' --resolution 1080p

# Custom output directory
youtube-downloader --url 'https://www.youtube.com/watch?v=dQw4w9WgXcQ' --output ~/Videos

# Show help
youtube-downloader --help

# Show version
youtube-downloader --version
```

### Flags

| Flag | Description | Default |
|---|---|---|
| `--url` | YouTube video URL (required) | — |
| `--resolution` | Video resolution | `720p` |
| `--output` | Directory to save the video | `~/Downloads` |
| `--help` | Show help | — |
| `--version` | Show version | — |

> If the URL contains `&`, wrap it in single quotes to avoid shell issues.

## Development

### Prerequisites

- Go 1.26+

### Running

```bash
go run ./cmd/youtube-downloader/ --url 'https://www.youtube.com/watch?v=dQw4w9WgXcQ'
```

### Tests

```bash
go test ./...
```

### Build

```bash
go build -o bin/youtube-downloader ./cmd/youtube-downloader/
```

## Project structure

```
cmd/youtube-downloader/    # Entry point (main)
internal/
  app/                     # Use cases (business logic)
  cli/                     # Argument parser and terminal utilities
  downloader/              # Download service interface
  fs/                      # File system operations
  youtube/                 # YouTube API client
```

The project follows Clean Architecture and SOLID principles, with dependency inversion across layers.
