#!/usr/bin/env bash
# dotenvx-tui — Interactive terminal UI for managing dotenvx-encrypted env vars.
# Requires: gum (brew install gum), dotenvx, jq
#
# Usage:
#   bash env-manage.sh [TARGET_DIR]
#
# TARGET_DIR defaults to the current working directory.
# The script auto-discovers encrypted .env.* files in the project tree.
set -euo pipefail

# Use the first argument as target dir, or fall back to cwd
TARGET_DIR="${1:-$(pwd)}"
TARGET_DIR="$(cd "$TARGET_DIR" && pwd)"
cd "$TARGET_DIR"

# ─── Dependency check ────────────────────────────────────────────────────────
for cmd in gum dotenvx jq; do
  if ! command -v "$cmd" &>/dev/null; then
    echo "Missing dependency: $cmd" >&2
    [[ "$cmd" == "gum" ]] && echo "  Install: brew install gum" >&2
    [[ "$cmd" == "jq" ]] && echo "  Install: brew install jq" >&2
    exit 1
  fi
done

# ─── Discovery ───────────────────────────────────────────────────────────────
# Find all dotenvx-encrypted .env.* files in the project tree.
# Encrypted files contain a DOTENV_PUBLIC_KEY header line.
DISCOVERED_FILES=()
while IFS= read -r -d '' file; do
  rel="${file#$TARGET_DIR/}"
  DISCOVERED_FILES+=("$rel")
done < <(
  find "$TARGET_DIR" -name '.env.*' -type f \
    -not -name '.env.keys' \
    -not -name '*.example' \
    -not -name '.envrc' \
    -not -name '.env.vault' \
    -print0 \
  | while IFS= read -r -d '' f; do
      grep -q 'DOTENV_PUBLIC_KEY' "$f" 2>/dev/null && printf '%s\0' "$f"
    done
)

if [ ${#DISCOVERED_FILES[@]} -eq 0 ]; then
  echo "No dotenvx-encrypted .env files found in $TARGET_DIR" >&2
  echo "Encrypted files must contain a DOTENV_PUBLIC_KEY header (created by dotenvx encrypt)." >&2
  exit 1
fi

# ─── Derive scopes and environments ──────────────────────────────────────────
# Scope = directory containing the env file, relative to TARGET_DIR ("." for root).
# Environment = everything after ".env." in the filename.

get_scopes() {
  for f in "${DISCOVERED_FILES[@]}"; do
    dirname "$f"
  done | sort -u
}

get_envs_for_scope() {
  local scope="$1"
  for f in "${DISCOVERED_FILES[@]}"; do
    local dir
    dir="$(dirname "$f")"
    if [[ "$dir" == "$scope" ]]; then
      basename "$f" | sed 's/^\.env\.//'
    fi
  done | sort -u
}

# ─── Helpers ─────────────────────────────────────────────────────────────────
resolve_file() {
  local scope="$1" env="$2"
  if [[ "$scope" == "." ]]; then
    echo ".env.$env"
  else
    echo "$scope/.env.$env"
  fi
}

display_label() {
  local scope="$1" env="$2"
  resolve_file "$scope" "$env"
}

get_decrypted_keys() {
  local file="$1"
  dotenvx get -f "$file" 2>/dev/null \
    | jq -r 'keys[]' 2>/dev/null \
    | grep -v '^DOTENV_PUBLIC_KEY' \
    | sort
}

# ─── Scope selection ─────────────────────────────────────────────────────────
mapfile -t SCOPES < <(get_scopes)

if [ ${#SCOPES[@]} -eq 1 ]; then
  SCOPE="${SCOPES[0]}"
else
  SCOPE=$(printf '%s\n' "${SCOPES[@]}" | gum choose --header "Which scope?")
fi

# ─── Action ──────────────────────────────────────────────────────────────────
ACTION=$(gum choose --header "What do you want to do?" \
  "set — add or rotate a secret" \
  "get — view a decrypted value" \
  "list — show all vars in a file" \
  "diff — compare two environments")
ACTION="${ACTION%% —*}"

# ─── Environment selection helper ────────────────────────────────────────────
mapfile -t AVAILABLE_ENVS < <(get_envs_for_scope "$SCOPE")

pick_env() {
  local header="${1:-Which environment?}"
  if [ ${#AVAILABLE_ENVS[@]} -eq 1 ]; then
    echo "${AVAILABLE_ENVS[0]}"
  else
    printf '%s\n' "${AVAILABLE_ENVS[@]}" | gum choose --header "$header"
  fi
}

# ─── Actions ─────────────────────────────────────────────────────────────────
case "$ACTION" in
  set)
    ENV=$(pick_env)
    FILE=$(resolve_file "$SCOPE" "$ENV")

    if [ ! -f "$FILE" ]; then
      echo "File not found: $FILE" >&2
      exit 1
    fi

    # Show existing keys for reference
    KEYS=$(get_decrypted_keys "$FILE" || true)

    if [ -n "$KEYS" ]; then
      echo ""
      gum style --foreground 242 "Existing vars in $(display_label "$SCOPE" "$ENV"):"
      echo "$KEYS" | gum style --foreground 242 --padding "0 2"
      echo ""
    fi

    # Pick existing key or type a new one
    KEY=$(echo -e "$KEYS\n— enter a new key —" \
      | gum filter --header "Select or search for a key:" --placeholder "type to filter...")

    if [ "$KEY" = "— enter a new key —" ] || [ -z "$KEY" ]; then
      KEY=$(gum input --header "New key name:" --placeholder "MY_NEW_VAR")
    fi

    if [ -z "$KEY" ]; then
      echo "No key provided." >&2
      exit 1
    fi

    # Show current value if it exists
    CURRENT=$(dotenvx get "$KEY" -f "$FILE" 2>/dev/null || true)
    if [ -n "$CURRENT" ]; then
      if [ "${#CURRENT}" -gt 4 ]; then
        MASKED="${CURRENT:0:4}$(printf '%*s' $((${#CURRENT} - 4)) '' | tr ' ' '•')"
      else
        MASKED="••••"
      fi
      gum style --foreground 214 "Current value: $MASKED"
    fi

    VALUE=$(gum input --header "New value for $KEY:" --placeholder "paste or type the value")

    if [ -z "$VALUE" ]; then
      echo "No value provided." >&2
      exit 1
    fi

    dotenvx set "$KEY" "$VALUE" -f "$FILE"
    echo ""
    gum style --foreground 34 "✓ Set $KEY in $(display_label "$SCOPE" "$ENV")"
    echo ""
    gum style --foreground 242 "Next: git add $FILE && git commit -m 'chore(env): rotate $KEY'"
    ;;

  get)
    ENV=$(pick_env)
    FILE=$(resolve_file "$SCOPE" "$ENV")

    KEYS=$(get_decrypted_keys "$FILE" || true)
    if [ -z "$KEYS" ]; then
      echo "No vars found or unable to decrypt $FILE" >&2
      echo "Ensure the matching .env.keys file exists alongside the env file." >&2
      exit 1
    fi

    KEY=$(echo "$KEYS" | gum filter --header "Select a key to reveal:" --placeholder "type to filter...")

    if [ -z "$KEY" ]; then
      echo "No key selected." >&2
      exit 1
    fi

    VALUE=$(dotenvx get "$KEY" -f "$FILE" 2>/dev/null || echo "(not set)")
    echo ""
    gum style --foreground 214 --bold "$KEY"
    echo "$VALUE"
    ;;

  list)
    ENV=$(pick_env)
    FILE=$(resolve_file "$SCOPE" "$ENV")

    KEYS=$(get_decrypted_keys "$FILE" || true)
    if [ -z "$KEYS" ]; then
      echo "No vars found or unable to decrypt $FILE" >&2
      echo "Ensure the matching .env.keys file exists alongside the env file." >&2
      exit 1
    fi

    echo ""
    gum style --bold "$(display_label "$SCOPE" "$ENV")"
    echo "$KEYS" | while read -r key; do
      echo "  $key"
    done
    ;;

  diff)
    if [ ${#AVAILABLE_ENVS[@]} -lt 2 ]; then
      echo "Need at least two environments to diff (found ${#AVAILABLE_ENVS[@]} in scope '$SCOPE')." >&2
      exit 1
    fi

    ENV1=$(pick_env "First environment:")
    ENV2=$(pick_env "Second environment:")

    if [ "$ENV1" = "$ENV2" ]; then
      echo "Cannot diff an environment against itself." >&2
      exit 1
    fi

    FILE1=$(resolve_file "$SCOPE" "$ENV1")
    FILE2=$(resolve_file "$SCOPE" "$ENV2")

    KEYS1=$(get_decrypted_keys "$FILE1" || true)
    KEYS2=$(get_decrypted_keys "$FILE2" || true)

    if [ -z "$KEYS1" ] || [ -z "$KEYS2" ]; then
      echo "Unable to decrypt one or both env files." >&2
      echo "Ensure the matching .env.keys file exists alongside each env file." >&2
      exit 1
    fi

    echo ""
    gum style --bold "Comparing $ENV1 vs $ENV2 in $SCOPE"
    echo ""

    # Keys only in env1
    ONLY1=$(comm -23 <(echo "$KEYS1") <(echo "$KEYS2") || true)
    if [ -n "$ONLY1" ]; then
      gum style --foreground 196 "Only in $ENV1:"
      echo "$ONLY1" | sed 's/^/  /'
      echo ""
    fi

    # Keys only in env2
    ONLY2=$(comm -13 <(echo "$KEYS1") <(echo "$KEYS2") || true)
    if [ -n "$ONLY2" ]; then
      gum style --foreground 34 "Only in $ENV2:"
      echo "$ONLY2" | sed 's/^/  /'
      echo ""
    fi

    # Shared keys with different values
    SHARED=$(comm -12 <(echo "$KEYS1") <(echo "$KEYS2") || true)
    DIFF_KEYS=()
    SAME_COUNT=0
    while IFS= read -r key; do
      [ -z "$key" ] && continue
      v1=$(dotenvx get "$key" -f "$FILE1" 2>/dev/null || true)
      v2=$(dotenvx get "$key" -f "$FILE2" 2>/dev/null || true)
      if [ "$v1" != "$v2" ]; then
        DIFF_KEYS+=("$key")
      else
        SAME_COUNT=$((SAME_COUNT + 1))
      fi
    done <<< "$SHARED"

    if [ ${#DIFF_KEYS[@]} -gt 0 ]; then
      gum style --foreground 214 "Different values:"
      for k in "${DIFF_KEYS[@]}"; do
        echo "  $k"
      done
      echo ""
    fi

    if [ "$SAME_COUNT" -gt 0 ]; then
      gum style --foreground 242 "$SAME_COUNT vars identical across both environments"
    fi
    ;;
esac
