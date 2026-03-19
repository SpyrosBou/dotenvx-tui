# Changelog

All notable changes to this project will be documented in this file.

## [2.1.0] - 2026-03-19

### Added
- Create new variables with `n` keybinding (opens set overlay in new-key mode)
- Delete variables with `x` keybinding (decrypt, remove, re-encrypt with confirmation)
- Multi-select delete support (select keys with space, then `x` to delete all)

### Fixed
- Pressing `q` during text input in overlays no longer quits the app
- Import overlay now steps back from key selection to file picker on escape
- Escape in Keys panel hides revealed preview before moving focus

## [2.0.0] - 2026-03-17

Complete rewrite from bash script to full-screen Go TUI.

### Distribution
- Published to npm: `npm install -g dotenvx-tui` or `npx dotenvx-tui`
- Automated releases via GoReleaser (cross-compiled binaries + npm + GitHub Releases)

### Added
- Three-panel browser layout (Scopes | Environments | Keys)
- Keyboard-driven navigation with circular wrapping (j/k, Tab/Shift+Tab)
- Multi-select batch operations (Space to toggle, `a` for select all)
- Set value overlay with real-time key name validation and batch mode
- Environment diff with color-coded comparison (only-left, only-right, different, identical)
- Bulk import from plaintext `.env` files with key-level selection
- Export to clipboard in `KEY=VALUE` format
- Copy single or multiple decrypted values to clipboard
- Live file watching via fsnotify (auto-refresh on disk changes)
- Help overlay with full keybinding reference (`?`)
- Preview pane with masked/revealed values and auto-mask timer (30s)
- Responsive layout (collapses scopes panel on narrow terminals)
- `--help` and `--version` CLI flags
- Empty state and fatal error views with setup instructions

### Security
- `SecureBytes` type that zeros memory on cleanup
- Key name validation (`[A-Za-z_][A-Za-z0-9_]*`, rejects `DOTENV_` prefix)
- File path traversal prevention with symlink escape detection
- Minimal subprocess environment (PATH/HOME/TERM only)
- Decrypted values never included in error messages
- Values masked by default (first 4 chars + bullets)
- `exec.Command` with explicit arguments (no shell interpolation)

### Removed
- `gum` dependency (replaced by Bubbletea)
- `jq` dependency (replaced by native JSON parsing)

### Changed
- Rewritten from bash to Go using Bubbletea v2, Bubbles v2, and Lipgloss v2
- Original bash script moved to `legacy/env-manage.sh`

## [1.0.0] - 2026-03-17

Initial release as a bash script.

### Added
- Interactive CLI for managing dotenvx-encrypted env vars
- Four actions: set, get, list, diff
- Auto-discovery of `.env.*` files with `DOTENV_PUBLIC_KEY` validation
- Support for flat projects, monorepos, and arbitrary directory depths
- Scope and environment derivation from file paths
