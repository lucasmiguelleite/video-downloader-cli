# video-downloader

A Go CLI to download videos from YouTube and Kick from the command line.

## Installation

### Via `go install`

```bash
go install ./cmd/video-downloader/
```

The binary becomes available at `$GOPATH/bin` (or `$HOME/go/bin`).

### Via `go build`

```bash
go build -o bin/video-downloader ./cmd/video-downloader/
```

## Prerequisites

- **Go 1.26+**
- **ffmpeg** — required for Kick VOD downloads (converts HLS `.ts` segments to `.mp4`)
  ```bash
  brew install ffmpeg
  ```

## Usage

```bash
# YouTube
video-downloader --url 'https://www.youtube.com/watch?v=dQw4w9WgXcQ'

# Kick VOD
video-downloader --url 'https://kick.com/nitrao/videos/7b0877f1-983d-4c5f-932e-cda88e8196f1'

# Specify resolution
video-downloader --url 'https://www.youtube.com/watch?v=dQw4w9WgXcQ' --resolution 1080p

# Custom output directory
video-downloader --url 'https://youtube.com/watch?v=xxx' --output ~/Videos

# Show help
video-downloader --help

# Show version
video-downloader --version
```

### Flags

| Flag | Description | Default |
|---|---|---|
| `--url` | Video URL — supports YouTube and Kick (required) | — |
| `--resolution` | Video resolution | `720p` |
| `--output` | Directory to save the video | `~/Downloads` |
| `--help` | Show help | — |
| `--version` | Show version | — |

> If the URL contains `&`, wrap it in single quotes to avoid shell issues.

## Development

### Running

```bash
go run ./cmd/video-downloader/ --url 'https://www.youtube.com/watch?v=dQw4w9WgXcQ'
```

### Tests

```bash
go test ./...
```

### Build

```bash
go build -o bin/video-downloader ./cmd/video-downloader/
```

## Project structure

```
cmd/video-downloader/    # Entry point (main)
internal/
  app/                   # Use cases (business logic)
  cli/                   # Argument parser and terminal utilities
  downloader/            # Download service interface
  fs/                    # File system operations
  kick/                  # Kick VOD client (HLS/m3u8)
  youtube/               # YouTube API client
```

The project follows Clean Architecture and SOLID principles, with dependency inversion across layers.
