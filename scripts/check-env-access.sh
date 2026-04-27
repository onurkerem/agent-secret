#!/usr/bin/env bash
set -euo pipefail

INPUT=$(cat)

TOOL_NAME=$(echo "$INPUT" | jq -r '.tool_name // .toolName // empty')

case "$TOOL_NAME" in
  Read|ReadFile)
    TARGET=$(echo "$INPUT" | jq -r '.tool_input.file_path // .tool_input.path // .tool_input.filePath // empty')
    ;;
  Bash|Shell)
    TARGET=$(echo "$INPUT" | jq -r '.tool_input.command // .tool_input.cmd // empty')
    ;;
  Grep|rg)
    TARGET=$(echo "$INPUT" | jq -r '.tool_input.path // .tool_input.file_path // empty')
    ;;
  Glob)
    PATTERN=$(echo "$INPUT" | jq -r '.tool_input.pattern // .tool_input.glob_pattern // empty')
    PATH_VAL=$(echo "$INPUT" | jq -r '.tool_input.path // .tool_input.target_directory // empty')
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
  if { [ "$TOOL_NAME" = "Bash" ] || [ "$TOOL_NAME" = "Shell" ]; } && echo "$TARGET" | grep -q 'agent-secret'; then
    exit 0
  fi
  echo "You are trying to access secrets you are not allowed to. Use agent-secret tool or ask for help from user. Do not try to overcome this measure." >&2
  exit 2
fi

exit 0
