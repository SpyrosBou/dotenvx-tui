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
 s:set  g:get  d:diff  i:import  e:export  c:copy
```

## Install

```bash
npm install -g dotenvx-tui
```

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

## Features

- Three-panel browser (Scopes | Environments | Keys)
- Keyboard-driven navigation with circular wrapping
- Multi-select batch operations
- Set, get, diff, import, export, and copy to clipboard
- Live file watching (auto-refresh on changes)
- Masked value preview with auto-mask timer
- Memory-safe secret handling

Full documentation: [github.com/warui1/dotenvx-tui](https://github.com/warui1/dotenvx-tui)

## License

MIT
