#!/usr/bin/env bash
set -euo pipefail

BOLD='\033[1m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
RESET='\033[0m'

info()  { echo -e "${GREEN}✓${RESET} $*"; }
warn()  { echo -e "${YELLOW}!${RESET} $*"; }
error() { echo -e "${RED}✗${RESET} $*" >&2; exit 1; }

HOOK_SCRIPT_DIR="$HOME/.agent-secret/hooks"
HOOK_SCRIPT_PATH="$HOOK_SCRIPT_DIR/check-env-access.sh"
HOOK_SCRIPT_URL="https://raw.githubusercontent.com/onurkerem/agent-secret/main/scripts/check-env-access.sh"
CLAUDE_SETTINGS=".claude/settings.json"
CODEX_HOOKS=".codex/hooks.json"

echo ""
echo -e "${BOLD}  agent-secret — Secure Local Secret Vault${RESET}"
echo -e "  ${BOLD}One-liner installer${RESET}"
echo ""

# --- OS check ---
OS="$(uname -s)"
case "$OS" in
  Darwin|Linux) ;;
  *) error "Unsupported OS: $OS. This installer requires macOS or Linux." ;;
esac

# --- Prerequisites ---
command -v curl >/dev/null 2>&1 || error "curl is required but not found. Install it first."

if ! command -v jq >/dev/null 2>&1; then
  if command -v brew >/dev/null 2>&1; then
    info "Installing jq via Homebrew..."
    brew install jq
  else
    error "jq is required but not found. Install it: https://jqlang.github.io/jq/download/"
  fi
fi
info "Prerequisites met (curl, jq)"

# --- Install agent-secret ---
if command -v agent-secret >/dev/null 2>&1; then
  info "agent-secret already installed: $(agent-secret --version 2>/dev/null || echo 'unknown version')"
elif command -v brew >/dev/null 2>&1; then
  info "Installing agent-secret via Homebrew..."
  brew tap onurkerem/agent-secret
  brew install agent-secret
  info "Installed: $(agent-secret --version 2>/dev/null || echo 'done')"
elif command -v go >/dev/null 2>&1; then
  info "Installing agent-secret via Go..."
  cd "$(mktemp -d)"
  git clone https://github.com/onurkerem/agent-secret.git
  cd agent-secret/packages/cli
  go build -o "$(go env GOPATH)/bin/agent-secret" .
  cd -
  rm -rf "$(pwd)/agent-secret"
  info "Installed: $(agent-secret --version 2>/dev/null || echo 'done')"
else
  error "Neither Homebrew nor Go found. Install one: https://brew.sh or https://go.dev/dl/"
fi

# --- Download hook script ---
info "Installing hook script to $HOOK_SCRIPT_PATH..."
mkdir -p "$HOOK_SCRIPT_DIR"
curl -fsSL "$HOOK_SCRIPT_URL" -o "$HOOK_SCRIPT_PATH"
chmod +x "$HOOK_SCRIPT_PATH"
info "Hook script installed"

# --- Set up Claude Code hooks ---
if [ -f "$CLAUDE_SETTINGS" ]; then
  EXISTING=$(jq '.hooks.PreToolUse // [] | map(.matcher)' "$CLAUDE_SETTINGS" 2>/dev/null || echo '[]')
  if echo "$EXISTING" | jq -e 'index("Read|Bash|Grep|Glob")' >/dev/null 2>&1; then
    info "Claude Code hooks already configured in $CLAUDE_SETTINGS"
  else
    NEW_HOOK='{"matcher":"Read|Bash|Grep|Glob","hooks":[{"type":"command","command":"'"$HOOK_SCRIPT_PATH"'"}]}'
    jq --argjson hook "$NEW_HOOK" '
      .hooks = (.hooks // {}) |
      .hooks.PreToolUse = (.hooks.PreToolUse // []) + [$hook]
    ' "$CLAUDE_SETTINGS" > "${CLAUDE_SETTINGS}.tmp" && mv "${CLAUDE_SETTINGS}.tmp" "$CLAUDE_SETTINGS"
    info "Added Claude Code hooks to $CLAUDE_SETTINGS"
  fi
else
  mkdir -p .claude
  cat > "$CLAUDE_SETTINGS" <<HOOKJSON
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Read|Bash|Grep|Glob",
        "hooks": [
          {
            "type": "command",
            "command": "$HOOK_SCRIPT_PATH"
          }
        ]
      }
    ]
  }
}
HOOKJSON
  info "Created Claude Code hooks at $CLAUDE_SETTINGS"
fi

# --- Set up Codex hooks ---
if [ -f "$CODEX_HOOKS" ]; then
  EXISTING=$(jq '.hooks.PreToolUse // [] | map(.matcher)' "$CODEX_HOOKS" 2>/dev/null || echo '[]')
  if echo "$EXISTING" | jq -e 'index("Bash")' >/dev/null 2>&1; then
    info "Codex hooks already configured in $CODEX_HOOKS"
  else
    NEW_HOOK='{"matcher":"Bash","hooks":[{"type":"command","command":"'"$HOOK_SCRIPT_PATH"'"}]}'
    jq --argjson hook "$NEW_HOOK" '
      .hooks = (.hooks // {}) |
      .hooks.PreToolUse = (.hooks.PreToolUse // []) + [$hook]
    ' "$CODEX_HOOKS" > "${CODEX_HOOKS}.tmp" && mv "${CODEX_HOOKS}.tmp" "$CODEX_HOOKS"
    info "Added Codex hooks to $CODEX_HOOKS"
  fi
else
  mkdir -p .codex
  cat > "$CODEX_HOOKS" <<HOOKJSON
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "$HOOK_SCRIPT_PATH"
          }
        ]
      }
    ]
  }
}
HOOKJSON
  info "Created Codex hooks at $CODEX_HOOKS"
fi

# --- Codex feature flag ---
CODEX_CONFIG="$HOME/.codex/config.toml"
if [ -f "$CODEX_CONFIG" ]; then
  if ! grep -q 'codex_hooks' "$CODEX_CONFIG"; then
    echo "" >> "$CODEX_CONFIG"
    echo "[features]" >> "$CODEX_CONFIG"
    echo "codex_hooks = true" >> "$CODEX_CONFIG"
    info "Enabled codex_hooks feature in $CODEX_CONFIG"
  elif grep -q 'codex_hooks = false' "$CODEX_CONFIG"; then
    sed -i'' 's/codex_hooks = false/codex_hooks = true/' "$CODEX_CONFIG"
    info "Enabled codex_hooks feature in $CODEX_CONFIG"
  fi
fi

# --- Summary ---
echo ""
echo -e "${BOLD}  Installation complete!${RESET}"
echo ""
echo "  Verify:  agent-secret --version"
echo ""
echo -e "  ${YELLOW}To install the AI agent skill, run:${RESET}"
echo "    npx skills add https://github.com/onurkerem/agent-secret --skill agent-secret"
echo ""
