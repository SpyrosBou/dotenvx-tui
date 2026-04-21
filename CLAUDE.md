# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

See `AGENTS.md` for the full contributor guide (structure, style, commit scopes). This file covers the pieces that are easier to get wrong than to discover.

## Commands

- `go run . [path]` — run the TUI locally against `path` (defaults to cwd). `dotenvx` must be on `PATH`.
- `make build` — build the local `dotenvx-tui` binary; `VERSION` is injected via `-X main.version=...` from `git describe`.
- `make test` / `go test ./...` — run all Go tests.
- Single test: `go test ./internal/validate -run TestKeyName` (tests live beside code as `*_test.go` under `internal/`).
- `make lint` — `golangci-lint run ./...` (install `golangci-lint` locally first; the Makefile does not bootstrap it).
- `make build-all` — cross-compile static darwin/linux × arm64/amd64 binaries into `dist/`.
- `make release-dry-run` — `goreleaser release --snapshot --clean` (validates `.goreleaser.yaml` without publishing).
- `./scripts/set-version.sh <version>` — update `npm/package.json` version (called by the release workflow; don't hand-edit the version).
- `gh workflow run npm-publish.yml -f version=<version>` — publish the npm package independently of the GoReleaser GitHub release workflow. It packages from `main` by default; pass `-f ref=<tag-or-branch>` when exact source matching matters.

The committed `./dotenvx-tui` binary in the repo root is gitignored — `make build` overwrites it.

## Architecture

### Entry point and Bubble Tea root

`main.go` parses flags, resolves the target directory, and starts `tea.NewProgram(ui.NewModel(targetDir))`. Version is a package-level `var version = "dev"` overridden by ldflags.

The Bubble Tea root lives in `internal/ui/`: `model.go` (`Model` struct), `update.go` (`Update`), `view.go` (`View`), `layout.go` (panel sizing), `keys.go` (`DefaultKeyMap`), `messages.go` (all message types). Uses `charm.land/bubbletea/v2` + `charm.land/lipgloss/v2` + `charm.land/bubbles/v2`.

### Update ordering — important

`Update` does three things in a specific order, and it must stay in this order:

1. `ctrl+c` always quits (even during overlays).
2. **Overlay *result* messages** (`overlays.SetDoneMsg`, `ImportDoneMsg`, `DeleteDoneMsg`, etc.) are handled *before* dispatching to the active overlay. They arrive asynchronously and the overlay may already be closed — routing them through the overlay's `Update` loses the status line. (See commit `0ffa024` for the regression this fixed.)
3. Then, if an overlay is active, `updateOverlay` owns the message. Otherwise it falls through to window/background/discovery/key handlers.

`WatcherFileChangedMsg` is also handled at the top level: it triggers a re-discovery and re-subscribes via `m.fileWatcher.Cmd()`. If you stop re-subscribing, the watcher only fires once.

### Panel cascade

Three-panel focus model: `PanelScopes → PanelEnvs → PanelKeys`. Selecting a scope repopulates envs and auto-selects the first; selecting an env loads keys; moving the cursor in the keys panel triggers `loadValue` to refresh the preview. Scope is `"."` for files at the target dir root. When there's only one scope and width < 100, the scopes panel is hidden (`Layout.HideScopes`) and focus skips it.

### Discovery

`dotenvx.Discover` walks from `targetDir`, skipping `node_modules`, `.git`, `vendor`. A file is considered a dotenvx-encrypted env file only if:

- Name starts with `.env.`
- Not in the excluded set (`.env.keys`, `.env.vault`, `.envrc`)
- Doesn't end in `.example` or `.sample`
- **Contains `DOTENV_PUBLIC_KEY` in its first 20 lines** (`HasPublicKeyHeader`)

The header check is what separates real dotenvx files from plaintext `.env.local`. If you add new exclusion rules, update both `Discover` and the README section "Excluded from discovery".

### dotenvx subprocess boundary

`internal/dotenvx/runner.go` is the only place that shells out to `dotenvx`. Rules enforced here and expected to stay that way:

- `exec.LookPath("dotenvx")` at `NewRunner` time; the resolved absolute path is cached.
- Every call uses `exec.CommandContext` with a 30s timeout.
- Every call uses `minimalEnv()` (PATH, HOME, TERM only). **Never pass `os.Environ()`** — it would leak the parent process's secrets into dotenvx's subprocess.
- Every call first runs `validate.FilePath(workDir, file)` to prevent path traversal and symlink escape.
- `stderr` bytes are the source of error messages; **never surface `stdout`** in errors (it contains decrypted values).
- `stdout` buffers holding decrypted data are zeroed with a `for i := range ...` loop before the function returns.
- `Set` and `Unset` do not call `dotenvx set` or decrypt the target file in place. They use `decrypt --stdout`, transform plaintext in memory, write a private staged plaintext file, run `encrypt --stdout`, and atomically replace the encrypted target.
- The staged plaintext file is a compromise forced by dotenvx's file-oriented encrypt command. It must stay `0600`, in a temp directory, and removed with `defer os.RemoveAll`.

`formatEnvValue` intentionally rejects values with newlines or values containing all three quote delimiters (backtick, single quote, and double quote). Do not silently fall back to passing those values through process arguments; that reintroduces the original leak.

### Secret handling (`internal/secret/secure.go`)

Decrypted values live in `*SecureBytes`, never in Go strings. Key rules:

- `secret.New(src)` copies `src`, then zeroes the caller's `src` — ownership is transferred. Callers must not reuse `src`.
- `SecureBytes.Clear()` is mutex-guarded and idempotent; a finalizer calls it on GC. `Clear` is called on preview replacement, env/scope change, discovery refresh, and `cleanup()` on quit.
- `Masked()` shows first 4 chars + bullets; values ≤ 4 bytes are fully masked.
- `String()` exists but is a foot-gun — prefer `Read(fn)` or `Masked()`. Strings are immutable and cannot be zeroed.

### Key name & path validation

`internal/validate/validate.go`:

- `KeyName`: `^[A-Za-z_][A-Za-z0-9_]*$`, rejects `DOTENV_` prefix (reserved by dotenvx).
- `FilePath`: checks both `filepath.Abs` containment AND `filepath.EvalSymlinks` containment inside `workDir`.

### File watcher (`internal/watcher/watcher.go`)

fsnotify with a **100ms debounce**, filtered to `.env.*` writes/creates/removes/renames (directory events pass through so new subdirs get watched). Emits `FileChangedMsg` exactly once per debounce window. The model re-subscribes by calling `m.fileWatcher.Cmd()` after each message — forget this and refreshes stop after the first change.

`watchDirs` walks the target dir and adds every non-ignored directory individually (fsnotify doesn't recurse).

### Error types (`internal/apperr`)

Typed `AppError` with `Kind` enum for user-facing errors. `Message` must never contain decrypted values — if you're wrapping a runner error, only `err.Error()` (which is sanitized stderr) is safe.

## npm distribution

`npm/` is a thin Node wrapper. `npm/install.js` is a no-op (kept to not break existing installs); `npm/bin/cli.js` downloads the platform-specific binary from the GitHub release on first run, verifies against `package.json` `version`, caches in `npm/bin/dotenvx-tui`, then `spawn`s it. The downloaded binary is gitignored.

The npm launcher must download from `SpyrosBou/dotenvx-tui`, because that is where GoReleaser publishes release assets.

## Release pipeline

Push a `v*` tag → `.github/workflows/release.yml`:

1. `go test ./...`
2. GoReleaser builds cross-platform archives and publishes a GitHub release (owner `SpyrosBou` per `.goreleaser.yaml`).
3. `scripts/set-version.sh` updates `npm/package.json`.
4. `npm publish --access public` from `npm/`.

`NPM_TOKEN` must be a valid npm granular token for the `spyrosbou` maintainer with publish/write access to `dotenvx-tui`. If npm publish needs to be rerun after the GitHub Release already exists, do **not** rerun the full release workflow: GoReleaser will fail because the release already exists. Correct the npm credentials, then run the npm-only workflow:

```bash
gh workflow run npm-publish.yml --repo SpyrosBou/dotenvx-tui -f version=<version>
```

Don't manually edit `npm/package.json`'s `version` before tagging — the workflows rewrite it to match the tag/input. The repository module path, npm launcher repository, npm package metadata, and GoReleaser owner should all point at `github.com/SpyrosBou/dotenvx-tui`.

## Files to leave alone

- `legacy/env-manage.sh` — preserved original bash implementation. Not wired into the Go build; don't delete it or port changes here back to it.
- `dist/`, `dotenvx-tui` (repo root), `npm/bin/dotenvx-tui` — build artifacts, gitignored.
