# AGENTS.md

## Go

* Target Go 1.26+ (managed via `mise`; toolchain pinned in `mise.toml`).
* Follow idiomatic Go: effective, simple, readable. Prefer composition over inheritance, interfaces over concrete types, and small packages over god-packages.
* `cmd/` holds entry points; `internal/` holds all importable-by-binary-only packages. Nothing outside `internal/` should be imported by another module.
* Layout: `cmd/video-downloader/` (main), `internal/app` (use cases), `internal/cli` (arg parsing, terminal, logging), `internal/downloader` (domain contract: `Service`, `Video`, `RetryConfig`), `internal/hls` (m3u8 parsing, shared `FetchSegments`, resume `Manifest`), `internal/fs` (file + ffmpeg), and one package per provider: `internal/{youtube,twitch,kick}`.
* Each provider implements `downloader.Service` (`GetVideo`, `Download`) and exposes a `*API` interface (`client_interface.go`) with a real implementation (`api_real.go`). Tests define a local unexported `mockAPI` in the same package — no `api_fake.go` files.
* Accept `context.Context` as the first parameter of every function that does I/O or long work. Thread it end-to-end; never use `context.Background()` inside library code (only in `main`).
* Return errors from the deepest layer; wrap with `fmt.Errorf("...: %w", err)` at each boundary. Define sentinel errors (`errors.New`) for distinct failure modes and match with `errors.Is`.

## Code style

* Functions: 4–20 lines. Split if longer.
* Files: under 500 lines. Split by responsibility.
* One thing per function, one responsibility per package (SRP).
* Use specific names. Avoid `data`, `handler`, `manager`, `utils`, `helpers`, `service` — prefer the concrete noun (`FetchSegments`, `Manifest`, `DownloadUseCase`).
* No unused parameters, no unused imports. `go vet` must be clean.
* Prefer early returns. Max 2 levels of indentation (extract a helper when nesting deepens).
* Error messages must include the offending value and expected shape — e.g. `fmt.Errorf("twitch playlist fetch returned HTTP %d for %s", resp.StatusCode, m3u8URL)`.
* Don't duplicate logic across providers: if Twitch and Kick share a loop, it belongs in `internal/hls`. The shared `FetchSegments` helper is the canonical example.
* Avoid `init()`. Wire dependencies explicitly in `main`.

## Comments

* Keep existing comments during refactors.
* Explain why, not what.
* Add a doc comment to every exported function, type, and constant, starting with the identifier's name. State intent and include one usage example when non-obvious.
* Reference issue numbers or commit SHAs when code exists because of a specific bug or upstream constraint.

## Tests

* Run the whole suite with one command: `go test ./...` (or `go test -race ./...` for concurrency-sensitive code).
* New behavior is written test-first (TDD): write the failing test, watch it fail, implement to make it pass, refactor.
* Every new function gets a test. Every bug fix gets a regression test.
* Same-package tests (so unexported state is reachable). One `func TestXxx(t *testing.T)` per case is the project idiom; use table-driven tests only when a function genuinely has many input variants.
* Mock external I/O (HTTP, filesystem, ffmpeg) at the package's `*API` interface boundary with a local `mockAPI`/`mockService` struct — never hit the real network in tests.
* Tests should be F.I.R.S.T.: Fast (use `t.TempDir()`, `time.Millisecond` backoffs), Isolated, Repeatable, Self-validating (`t.Errorf`, not `fmt.Println`), Timely.
* Add `t.Cleanup` to restore any global state a test mutates (e.g. `downloader.DefaultRetryConfig`).

## Dependencies

* Stdlib-first. The only direct third-party deps are `kkdai/youtube/v2`, `schollz/progressbar/v3`, and `golang.org/x/sys`. Prefer the standard library (`net/http`, `log/slog`, `encoding/json`, `os/exec`) over adding a new module.
* Inject dependencies through function parameters or constructor functions (`NewClient(api)`, `NewDownloadUseCase(s)`). Never reach for a package-level variable from inside library code.
* Wrap third-party libraries behind project-owned interfaces — e.g. `youtube.YouTubeAPI` wraps `kkdai/youtube/v2.Client` so the rest of the code never imports it.
* Adding a dependency requires justifying why no stdlib solution fits, and keeping it as a direct (not indirect) `require` in `go.mod`.

## Formatting

* Use `gofmt` (or `go fmt ./...`). Don't manually enforce formatting rules, don't reformat code unrelated to your change.
* `gofmt -l .` must return no files. Run `goimports` if available for import grouping.

## Logging

* Use structured logging via `log/slog` (`slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: level}))`). Logs are key=value text records.
* Build the logger in `main` (`cli.NewLogger(os.Stderr, input.Verbose)`) and inject it via `WithLogger`. Library packages default to `slog.Default()`.
* Write logs to **stderr** so stdout stays clean for piping. The progress bar also writes to stderr.
* Log lifecycle events (download starting/complete, resume detected, remux), not per-segment spam. Reserve `slog.LevelDebug` for verbose diagnostics.
* Use plain text (`fmt.Println`/`fmt.Fprintf`) only for user-facing CLI output in `cmd/video-downloader/main.go` (e.g. the final "Saved to:" line).

## Commits

* Use Conventional Commits prefixes (`feat:`, `fix:`, `chore:`, `docs:`, `refactor:`, `perf:`, `ci:`).
* Write the subject and body in English, imperative mood ("add", not "added").
* Keep the subject under 72 characters, no trailing period.
* Wrap the body at ~72 columns; explain what and why, not only what.
* Make commits atomic — one logical change per commit.
* Do not add a `Co-Authored-By` trailer or any AI attribution.
* Commit `go.mod` and `go.sum` together; keep them in sync (`go mod tidy` before committing).
* Never commit generated artifacts (`bin/`, `dist/`, build output) or the `actions-runner/` workspace checkout.
* Never run `git push` automatically; only push on an explicit request.

## Build & verification

* Toolchain is managed by `mise` — run `mise install` once, or ensure `go` is on `PATH`.
* Build: `go build ./...`. Binary: `go build -o bin/video-downloader ./cmd/video-downloader/`.
* Verify before committing: `go build ./... && go vet ./... && go test -race ./... && gofmt -l .` (all must pass clean).
* Release artifacts are built by GoReleaser (`.goreleaser.yml`) via the GitHub Actions workflow in `.github/`.
