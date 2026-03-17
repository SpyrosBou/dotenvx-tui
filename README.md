# dotenvx-tui

Full-screen terminal UI for managing [dotenvx](https://dotenvx.com)-encrypted environment variables.

Browse, set, diff, import, and export secrets across environments — with a three-panel layout, multi-select, and keyboard-driven navigation.

```
┌─ Scopes ────┐ ┌─ Envs ───┐ ┌─ Keys ──────────────┐
│ > apps/api  │ │ > local  │ │ [x] DATABASE_URL    │
│   apps/web  │ │   staging│ │ [ ] API_KEY         │
│   pkg/db    │ │   prod   │ │ [x] REDIS_URL       │
└─────────────┘ └──────────┘ └─────────────────────┘
┌─ Preview ────────────────────────────────────────┐
│ DATABASE_URL = post••••••••••••••••••            │
└──────────────────────────────────────────────────┘
 s:set  g:get  d:diff  i:import  e:export  c:copy
```

## Prerequisites

| Tool | Install |
|------|---------|
| [dotenvx](https://dotenvx.com) | `brew install dotenvx/brew/dotenvx` |
| [Go 1.24+](https://go.dev) | `brew install go` (build only) |

**No longer requires** `gum` or `jq` — the Go rewrite handles everything natively.

## Install

```bash
# From source
go install github.com/warui1/dotenvx-tui@latest

# Or clone and build
git clone https://github.com/warui1/dotenvx-tui
cd dotenvx-tui
make install
```

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
| `s` | Set value (batch mode if multi-selected) |
| `g` | Get / decrypt value |
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
Select multiple keys with Space, then apply actions to all selected keys at once. Set values sequentially, copy all to clipboard, or export selected keys.

### Environment Diff
Compare two environments within the same scope. Color-coded output shows keys only in left, only in right, different values, and identical keys.

### Bulk Import
Import keys from a plaintext `.env` file into an encrypted one. Select which keys to import with checkboxes before confirming.

### Export to Clipboard
Export decrypted key-value pairs to clipboard in `KEY=VALUE` format.

### Live File Watching
Panels auto-refresh when `.env` files change on disk (via fsnotify).

### Security
- Decrypted values wrapped in `SecureBytes` type that zeros memory on cleanup
- Values masked by default (first 4 chars visible + bullets)
- Auto-mask after 30 seconds of reveal
- Key name validation (`[A-Za-z_][A-Za-z0-9_]*`)
- Path traversal prevention
- Minimal subprocess environment (PATH/HOME/TERM only)
- No secrets in error messages

## Project Structure Discovery

On launch, the TUI scans the target directory for dotenvx-encrypted `.env.*` files (identified by the `DOTENV_PUBLIC_KEY` header). It derives **scopes** (directories) and **environments** (file suffixes) automatically.

### Excluded from discovery
- `.env.keys` — private key files
- `*.example` / `*.sample` — template files
- `.envrc` — direnv configuration
- `.env.vault` — dotenvx vault files
- Files without `DOTENV_PUBLIC_KEY` header

## License

MIT
