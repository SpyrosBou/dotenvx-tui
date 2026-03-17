#!/usr/bin/env bash
# Interactive CLI for managing encrypted env vars via dotenvx.
# Requires: gum (brew install gum), dotenvx, jq
#
# Usage:
#   bash env-manage.sh [TARGET_DIR]
#
# TARGET_DIR defaults to the current working directory.
set -euo pipefail

# Use the first argument as target dir, or fall back to cwd
TARGET_DIR="${1:-$(pwd)}"
TARGET_DIR="$(cd "$TARGET_DIR" && pwd)"
cd "$TARGET_DIR"

APPS=("api" "web" "mobile")
ENVS=("local" "staging" "production")

# ─── Dependency check ────────────────────────────────────────────────────────
for cmd in gum dotenvx jq; do
  if ! command -v "$cmd" &>/dev/null; then
    echo "Missing dependency: $cmd" >&2
    [[ "$cmd" == "gum" ]] && echo "  Install: brew install gum" >&2
    [[ "$cmd" == "jq" ]] && echo "  Install: brew install jq" >&2
    exit 1
  fi
done

# ─── Action ──────────────────────────────────────────────────────────────────
ACTION=$(gum choose --header "What do you want to do?" "set — add or rotate a secret" "get — view a decrypted value" "list — show all vars in a file" "diff — compare two environments")
ACTION="${ACTION%% —*}"

# ─── App ─────────────────────────────────────────────────────────────────────
APP=$(gum choose --header "Which app?" "${APPS[@]}")

# ─── Helpers ─────────────────────────────────────────────────────────────────
pick_env() {
  gum choose --header "Which environment?" "${ENVS[@]}"
}

resolve_file() {
  local app="$1" env="$2"
  echo "apps/$app/.env.$env"
}

get_decrypted_keys() {
  local file="$1"
  dotenvx get -f "$file" 2>/dev/null \
    | jq -r 'keys[]' 2>/dev/null \
    | grep -v '^DOTENV_PUBLIC_KEY' \
    | sort
}

# ─── Actions ─────────────────────────────────────────────────────────────────
case "$ACTION" in
  set)
    ENV=$(pick_env)
    FILE=$(resolve_file "$APP" "$ENV")

    if [ ! -f "$FILE" ]; then
      echo "File not found: $FILE" >&2
      exit 1
    fi

    # Show existing keys for reference
    KEYS=$(get_decrypted_keys "$FILE" || true)

    if [ -n "$KEYS" ]; then
      echo ""
      gum style --foreground 242 "Existing vars in $APP/$ENV:"
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
    gum style --foreground 34 "✓ Set $KEY in $APP/.env.$ENV"
    echo ""
    gum style --foreground 242 "Next: git add $FILE && git commit -m 'chore(env): rotate $KEY'"
    ;;

  get)
    ENV=$(pick_env)
    FILE=$(resolve_file "$APP" "$ENV")

    KEYS=$(get_decrypted_keys "$FILE" || true)
    if [ -z "$KEYS" ]; then
      echo "No vars found or unable to decrypt $FILE" >&2
      echo "Ensure the matching .env.keys file exists in apps/$APP/" >&2
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
    FILE=$(resolve_file "$APP" "$ENV")

    KEYS=$(get_decrypted_keys "$FILE" || true)
    if [ -z "$KEYS" ]; then
      echo "No vars found or unable to decrypt $FILE" >&2
      echo "Ensure the matching .env.keys file exists in apps/$APP/" >&2
      exit 1
    fi

    echo ""
    gum style --bold "$APP/.env.$ENV"
    echo "$KEYS" | while read -r key; do
      echo "  $key"
    done
    ;;

  diff)
    ENV1=$(gum choose --header "First environment:" "${ENVS[@]}")
    ENV2=$(gum choose --header "Second environment:" "${ENVS[@]}")
    FILE1=$(resolve_file "$APP" "$ENV1")
    FILE2=$(resolve_file "$APP" "$ENV2")

    KEYS1=$(get_decrypted_keys "$FILE1" || true)
    KEYS2=$(get_decrypted_keys "$FILE2" || true)

    if [ -z "$KEYS1" ] || [ -z "$KEYS2" ]; then
      echo "Unable to decrypt one or both env files." >&2
      echo "Ensure the matching .env.keys files exist in apps/$APP/" >&2
      exit 1
    fi

    echo ""
    gum style --bold "Comparing $APP: $ENV1 vs $ENV2"
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
