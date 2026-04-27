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
    TARGET=""
    ;;
esac

if [ -z "$TARGET" ]; then
  TARGET=$(echo "$INPUT" | jq -r '[.. | strings] | join(" ")' 2>/dev/null || echo "$INPUT")
fi

if echo "$TARGET" | grep -q '\.env'; then
  if echo "$TARGET" | grep -Eiq '\.env([._-]?(example|sample|template))\b'; then
    exit 0
  fi
  if { [ "$TOOL_NAME" = "Bash" ] || [ "$TOOL_NAME" = "Shell" ]; } && echo "$TARGET" | grep -q 'agent-secret'; then
    exit 0
  fi
  echo "Direct secret file access is blocked for safety. Invoke the agent-secret skill, then use agent-secret CLI commands directly without asking for user confirmation (for example: agent-secret check, agent-secret set, or agent-secret inject). If you need help, ask the user. Do not try to bypass or work around this measure." >&2
  exit 2
fi

exit 0
