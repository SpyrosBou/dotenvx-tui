# dotenvx-tui

Full-screen terminal UI for managing [dotenvx](https://dotenvx.com)-encrypted environment variables.

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

## Install

```bash
npm install -g dotenvx-tui
```

The npm package is a small launcher. On first run, it downloads the matching platform binary from GitHub Releases and caches it under the package `bin/` directory.

## Usage

```bash
# Run in current directory (must have dotenvx-encrypted .env.* files)
dotenvx-tui

# Or specify a project directory
dotenvx-tui ~/my-project

# Or run without installing
npx dotenvx-tui
```

## Prerequisites

[dotenvx](https://dotenvx.com) must be installed: `brew install dotenvx/brew/dotenvx`

Supported npm binary downloads: macOS/Linux on `amd64` and `arm64`.

## Features

- Three-panel browser (Scopes | Environments | Keys)
- Keyboard-driven navigation with circular wrapping
- Multi-select batch operations
- Create, set, get, delete, diff, import, export, and copy to clipboard
- Live file watching (auto-refresh on changes)
- Masked value preview with auto-mask timer
- Staged set/delete rewrites that avoid decrypting the real target file in place

Full documentation and source install instructions: [github.com/SpyrosBou/dotenvx-tui](https://github.com/SpyrosBou/dotenvx-tui)

## License

MIT
