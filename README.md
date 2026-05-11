# video-downloader

A Go CLI to download videos from YouTube, Twitch, and Kick from the command line.

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
- **ffmpeg** — required for Twitch and Kick VOD downloads (converts HLS `.ts` segments to `.mp4`)
  ```bash
  brew install ffmpeg
  ```

## Usage

```bash
# YouTube
video-downloader --url 'https://www.youtube.com/watch?v=dQw4w9WgXcQ'

# Twitch VOD
video-downloader --url 'https://www.twitch.tv/videos/1234567890'

# Kick VOD
video-downloader --url 'https://kick.com/nitrao/videos/7b0877f1-983d-4c5f-932e-cda88e8196f1'

# Specify resolution
video-downloader --url 'https://www.youtube.com/watch?v=dQw4w9WgXcQ' --resolution 1080p

# Custom output directory
video-downloader --url 'https://twitch.tv/videos/1234567890' --output ~/Videos

# Adjust concurrency for HLS streams
video-downloader --url 'https://twitch.tv/videos/1234567890' --concurrency 10

# Show help
video-downloader --help

# Show version
video-downloader --version
```

### Flags

| Flag | Description | Default |
|---|---|---|
| `--url` | Video URL — YouTube, Twitch, Kick (required) | — |
| `--resolution` | Video resolution | `720p` |
| `--output` | Directory to save the video | `~/Downloads` |
| `--concurrency` | Parallel segment downloads (Twitch/Kick) | `20` |
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
  hls/                   # Shared HLS/m3u8 parser
  kick/                  # Kick VOD client
  twitch/                # Twitch VOD client (GraphQL + HLS)
  youtube/               # YouTube API client
```

The project follows Clean Architecture and SOLID principles, with dependency inversion across layers.
