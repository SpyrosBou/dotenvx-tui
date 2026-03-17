# dotenvx-tui

Interactive terminal UI for managing [dotenvx](https://dotenvx.com)-encrypted environment variables.

Browse, read, set, and diff secrets across environments — without memorizing dotenvx flags or file paths.

## Prerequisites

| Tool | Install |
|------|---------|
| [dotenvx](https://dotenvx.com) | `brew install dotenvx/brew/dotenvx` |
| [gum](https://github.com/charmbracelet/gum) | `brew install gum` |
| [jq](https://jqlang.github.io/jq/) | `brew install jq` |

## Usage

```bash
# Run from inside a project
cd ~/my-project
bash /path/to/env-manage.sh

# Or pass the project path as an argument
bash /path/to/env-manage.sh ~/my-project

# Or if cloned and using pnpm
cd ~/projects/dotenvx-tui
pnpm start ~/my-project
```

## How it works

On launch, the script scans the target directory for dotenvx-encrypted `.env.*` files (identified by the `DOTENV_PUBLIC_KEY` header that dotenvx adds during encryption). It derives available **scopes** (directories) and **environments** (file suffixes) automatically — no configuration needed.

### Supported project structures

**Flat project** — env files at the root:
```
.env.local
.env.staging
.env.production
```
The scope prompt is skipped entirely since there's only one.

**Monorepo** — env files nested in workspaces:
```
apps/api/.env.local
apps/api/.env.staging
apps/web/.env.local
apps/web/.env.production
packages/db/.env.local
```
You'll be prompted to pick a scope first, then an environment within it.

**Any depth works** — the script uses recursive `find`, so `infra/deploy/.env.production` is just as valid.

## Actions

### set — Add or rotate a secret

Presents existing keys for the selected file, lets you pick one to update or create a new one. Shows the current value (masked) before prompting for the new value. Encrypts and stores via `dotenvx set`.

### get — View a decrypted value

Lists all keys in the selected file with a fuzzy-searchable filter. Decrypts and displays the selected value.

### list — Show all vars in a file

Displays all decrypted key names for a given scope and environment.

### diff — Compare two environments

Compares two environment files within the same scope and reports:
- Keys only in the first environment
- Keys only in the second environment
- Keys present in both but with different values
- Count of identical keys

## Excluded from discovery

The following files are ignored during scanning:
- `.env.keys` — private key files
- `*.example` — template files
- `.envrc` — direnv configuration
- `.env.vault` — dotenvx vault files
- Any `.env.*` file without a `DOTENV_PUBLIC_KEY` header (not encrypted by dotenvx)

## License

MIT
