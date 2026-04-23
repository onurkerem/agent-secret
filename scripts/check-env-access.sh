#!/usr/bin/env bash
set -euo pipefail

INPUT=$(cat)

TOOL_NAME=$(echo "$INPUT" | jq -r '.tool_name // empty')

case "$TOOL_NAME" in
  Read)
    TARGET=$(echo "$INPUT" | jq -r '.tool_input.file_path // empty')
    ;;
  Bash)
    TARGET=$(echo "$INPUT" | jq -r '.tool_input.command // empty')
    ;;
  Grep)
    TARGET=$(echo "$INPUT" | jq -r '.tool_input.path // empty')
    ;;
  Glob)
    PATTERN=$(echo "$INPUT" | jq -r '.tool_input.pattern // empty')
    PATH_VAL=$(echo "$INPUT" | jq -r '.tool_input.path // empty')
    TARGET="${PATTERN} ${PATH_VAL}"
    ;;
  *)
    exit 0
    ;;
esac

if [ -z "$TARGET" ]; then
  exit 0
fi

if echo "$TARGET" | grep -q '\.env'; then
  if echo "$TARGET" | grep -q 'example'; then
    exit 0
  fi
  if [ "$TOOL_NAME" = "Bash" ] && echo "$TARGET" | grep -q 'agent-secret'; then
    exit 0
  fi
  echo "You are trying to access secrets you are not allowed to. Use agent-secret tool or ask for help from user. Do not try to overcome this measure." >&2
  exit 2
fi

exit 0
