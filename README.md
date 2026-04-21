# dotenvx-tui

[![GitHub release](https://img.shields.io/github/v/release/SpyrosBou/dotenvx-tui)](https://github.com/SpyrosBou/dotenvx-tui/releases)
[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

Full-screen terminal UI for managing [dotenvx](https://dotenvx.com)-encrypted environment variables.

Browse, set, delete, diff, import, and export secrets across environments with a three-panel layout, multi-select, and keyboard-driven navigation.

```
┌─ Scopes ────┐ ┌─ Envs ───┐ ┌─ Keys ──────────────┐
│ > apps/api  │ │ > local  │ │ [x] DATABASE_URL    │
│   apps/web  │ │   staging│ │ [ ] API_KEY         │
│   pkg/db    │ │   prod   │ │ [x] REDIS_URL       │
└─────────────┘ └──────────┘ └─────────────────────┘
┌─ Preview ────────────────────────────────────────┐
│ DATABASE_URL = post••••••••••••••••••            │
└──────────────────────────────────────────────────┘
 n:new  s:set  g:get  x:delete  d:diff  i:import  e:export  c:copy
```

## Prerequisites

| Tool | Required for | Install |
|------|--------------|---------|
| [dotenvx](https://dotenvx.com) | Runtime encryption/decryption | `brew install dotenvx/brew/dotenvx` |
| Node.js 16+ | npm wrapper | `npm install -g dotenvx-tui` |
| Go 1.26+ | Source builds | `go install github.com/SpyrosBou/dotenvx-tui@latest` |

## Install

```bash
# Via npm wrapper
npm install -g dotenvx-tui

# Or run directly with npx
npx dotenvx-tui

# From source (requires Go 1.26+)
go install github.com/SpyrosBou/dotenvx-tui@latest
```

The npm package is a small launcher that downloads the matching binary from GitHub Releases on first run. Prebuilt binaries are available for macOS and Linux on `amd64`/`arm64`.

Release archives are also available from the [GitHub Releases page](https://github.com/SpyrosBou/dotenvx-tui/releases).

## Usage

```bash
# Run in current directory
dotenvx-tui

# Or specify a project directory
dotenvx-tui ~/my-project

# Show help
dotenvx-tui --help
```

## Keybindings

### Navigation

| Key | Action |
|-----|--------|
| `tab` / `shift+tab` | Switch panels (Scopes → Envs → Keys) |
| `j`/`k` or `↑`/`↓` | Move cursor (circular wrapping) |
| `h`/`l` | Previous/Next panel |
| `enter` | Select item / Reveal decrypted value |
| `esc` | Close overlay / Go back |

### Selection

| Key | Action |
|-----|--------|
| `space` | Toggle multi-select (Keys panel) |
| `a` | Select all / Deselect all |

### Actions

| Key | Action |
|-----|--------|
| `n` | Create a new variable |
| `s` | Set value (batch mode if multi-selected) |
| `g` | Get / decrypt value |
| `x` | Delete selected variable(s) |
| `d` | Diff two environments |
| `i` | Import from plaintext .env file |
| `e` | Export to clipboard |
| `c` | Copy value to clipboard |
| `?` | Help overlay |
| `q` / `ctrl+c` | Quit |

## Features

### Three-Panel Browser
Scopes (directories), Environments (file suffixes), and Keys are displayed simultaneously. Navigate between panels with Tab, within panels with j/k.

### Multi-Select Batch Operations
Select multiple keys with Space, then apply actions to all selected keys. Set values sequentially, delete selected variables, copy values, or export selected keys.

### Environment Diff
Compare two environments within the same scope. Color-coded output shows keys only in left, only in right, different values, and identical keys.

### Bulk Import
Import keys from a plaintext `.env` file into an encrypted one. Select which keys to import with checkboxes before confirming.

### Export to Clipboard
Export decrypted key-value pairs to clipboard in `KEY=VALUE` format.

### Live File Watching
Panels auto-refresh when `.env` files change on disk (via fsnotify).

## Security Model

- Previewed values are wrapped in `SecureBytes` and cleared when replaced, hidden, or on shutdown.
- Values are masked by default and auto-mask again 30 seconds after reveal.
- Key names are validated with `[A-Za-z_][A-Za-z0-9_]*`; `DOTENV_` names are reserved.
- Env file paths are validated against the selected target directory, including symlink resolution.
- dotenvx subprocesses receive a minimal environment (`PATH`, `HOME`, `TERM` only).
- Set and delete workflows use `dotenvx decrypt --stdout`, a private staged plaintext file, `dotenvx encrypt --stdout`, and atomic replacement. The real target file is not decrypted in place.
- Set values are not passed to `dotenvx set` as process arguments.

### Limitations

- Clipboard actions intentionally place decrypted secrets in the OS clipboard and transient Go strings.
- Export and multi-copy operations materialize `KEY=VALUE` text in memory before writing to the clipboard.
- The staged set/delete path briefly writes plaintext to a `0600` temp file because dotenvx does not expose a documented stdin encryption mode.
- Staged set rejects values containing newlines, or values containing backticks, single quotes, and double quotes together, rather than risking dotenv parsing corruption.

## Project Structure Discovery

On launch, the TUI scans the target directory for dotenvx-encrypted `.env.*` files identified by a `DOTENV_PUBLIC_KEY` header in the first 20 lines. It derives **scopes** from directories and **environments** from file suffixes automatically.

### Excluded from discovery
- `.env.keys` — private key files
- `*.example` / `*.sample` — template files
- `.envrc` — direnv configuration
- `.env.vault` — dotenvx vault files
- Files without `DOTENV_PUBLIC_KEY` header

Plaintext `.env` files are not shown as managed encrypted files, but the import flow can read plaintext `.env*` files that are not encrypted.

## Upgrading from v1

v2 is a complete rewrite in Go. The original bash script is preserved at `legacy/env-manage.sh` if you need it.

What changed:
- No more `gum` or `jq` dependencies — single Go binary
- Full-screen TUI replaces sequential prompts
- Persistent session — no more restarting after each action
- Multi-select batch operations
- Clipboard, import, export, and file watching are new
- Better default masking and narrower subprocess exposure

## License

MIT
