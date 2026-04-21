# Repository Guidelines

## Project Structure & Module Organization

This is a Go TUI for managing dotenvx-encrypted environment files, with a small npm wrapper.

- `main.go` contains CLI entrypoint logic and version wiring.
- `internal/ui/` contains the Bubble Tea model, update/view flow, layout, key bindings, panels, and overlays.
- `internal/dotenvx/` handles encrypted env file discovery and dotenvx subprocess calls.
- `internal/secret/`, `internal/validate/`, `internal/watcher/`, `internal/theme/`, and `internal/clipboard/` hold focused support packages.
- Tests live beside code as `*_test.go` files under `internal/`.
- `npm/` contains the npm package wrapper and binary launcher.
- `scripts/` contains release helpers; `legacy/` preserves the original shell implementation.

Generated binaries and release artifacts (`dotenvx-tui`, `dist/`) are ignored.

## Build, Test, and Development Commands

- `go run . [path]` runs the TUI locally against the current directory or a target project path.
- `make build` builds the local `dotenvx-tui` binary with version metadata.
- `make test` or `go test ./...` runs all Go tests.
- `make lint` runs `golangci-lint run ./...`; install `golangci-lint` locally first.
- `make build-all` creates static Darwin/Linux binaries in `dist/`.
- `make release-dry-run` validates the GoReleaser configuration without publishing.

The TUI expects `dotenvx` to be installed and available on `PATH`.

## Coding Style & Naming Conventions

Use standard Go formatting (`gofmt`) and keep package names short, lowercase, and domain-focused. Prefer small packages with explicit responsibilities. Keep Bubble Tea message, model, and update behavior consistent with existing `internal/ui` patterns.

Use exported identifiers only when needed across packages. Environment variable keys must follow the existing validation style: uppercase snake case is preferred, and `DOTENV_` names are reserved.

## Testing Guidelines

Use Go’s standard `testing` package. Name tests `TestXxx` and keep table-driven cases close to the behavior being validated, as in `internal/validate/validate_test.go`. Add tests for discovery, validation, secret cleanup, and subprocess boundary behavior when those areas change.

## Commit & Pull Request Guidelines

Follow the repository’s Conventional Commit style: `feat(ui): ...`, `fix(npm): ...`, `chore(release): ...`. Keep scopes specific to the touched area (`ui`, `build`, `runtime`, `npm`, `release`).

For pull requests, include a short behavior summary, commands run, and any user-visible TUI changes. Include screenshots or terminal captures only when layout, key handling, or visual state changes.

## Security & Configuration Tips

Never commit `.env*` files, decrypted values, private keys, or generated npm binaries. Preserve the existing secret-handling approach: mask values by default, avoid secrets in errors, and use `internal/secret.SecureBytes` for decrypted material.
