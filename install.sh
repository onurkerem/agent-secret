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
CURSOR_HOOKS=".cursor/hooks.json"

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
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  arm64)  ARCH="arm64" ;;
  *)      error "Unsupported architecture: $ARCH" ;;
esac
case "$OS" in
  Darwin) OS_NAME="darwin" ;;
  Linux)  OS_NAME="linux" ;;
esac

LATEST_VERSION=$(curl -fsSL https://api.github.com/repos/onurkerem/agent-secret/releases/latest | jq -r '.tag_name // .name' | sed 's/^v//')
if [ -z "$LATEST_VERSION" ]; then
  error "Could not determine latest version"
fi

CURRENT_VERSION=""
if command -v agent-secret >/dev/null 2>&1; then
  CURRENT_VERSION=$(agent-secret --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
fi

INSTALL_DIR="$HOME/.local/bin"

if [ "$CURRENT_VERSION" = "$LATEST_VERSION" ]; then
  info "agent-secret v${CURRENT_VERSION} already up to date"
else
  if [ -n "$CURRENT_VERSION" ]; then
    info "Upgrading agent-secret from v${CURRENT_VERSION} to v${LATEST_VERSION}..."
  else
    info "Downloading agent-secret v${LATEST_VERSION} for ${OS_NAME}/${ARCH}..."
  fi

  DOWNLOAD_URL="https://github.com/onurkerem/agent-secret/releases/download/v${LATEST_VERSION}/agent-secret_${LATEST_VERSION}_${OS_NAME}_${ARCH}.tar.gz"
  TMPDIR_DOWNLOAD=$(mktemp -d)
  curl -fsSL "$DOWNLOAD_URL" -o "$TMPDIR_DOWNLOAD/agent-secret.tar.gz"
  tar -xzf "$TMPDIR_DOWNLOAD/agent-secret.tar.gz" -C "$TMPDIR_DOWNLOAD"
  mkdir -p "$INSTALL_DIR"
  mv "$TMPDIR_DOWNLOAD/agent-secret" "$INSTALL_DIR/agent-secret"
  chmod +x "$INSTALL_DIR/agent-secret"
  rm -rf "$TMPDIR_DOWNLOAD"
  info "Installed agent-secret v${LATEST_VERSION} to $INSTALL_DIR/agent-secret"

  if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
    warn "Add $INSTALL_DIR to your PATH:"
    echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
  fi
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
  jq -n --arg cmd "$HOOK_SCRIPT_PATH" '{
    hooks: {
      PreToolUse: [{
        matcher: "Read|Bash|Grep|Glob",
        hooks: [{ type: "command", command: $cmd }]
      }]
    }
  }' > "$CLAUDE_SETTINGS"
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
  jq -n --arg cmd "$HOOK_SCRIPT_PATH" '{
    hooks: {
      PreToolUse: [{
        matcher: "Bash",
        hooks: [{ type: "command", command: $cmd }]
      }]
    }
  }' > "$CODEX_HOOKS"
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

# --- Set up Cursor hooks ---
if [ -f "$CURSOR_HOOKS" ]; then
  EXISTING=$(jq '.hooks.preToolUse // [] | map(.matcher)' "$CURSOR_HOOKS" 2>/dev/null || echo '[]')
  if echo "$EXISTING" | jq -e 'index("Shell|Read")' >/dev/null 2>&1; then
    info "Cursor hooks already configured in $CURSOR_HOOKS"
  else
    NEW_HOOK="{\"command\":\"$HOOK_SCRIPT_PATH\",\"matcher\":\"Shell|Read\"}"
    jq --argjson hook "$NEW_HOOK" '
      .version = (.version // 1) |
      .hooks = (.hooks // {}) |
      .hooks.preToolUse = (.hooks.preToolUse // []) + [$hook]
    ' "$CURSOR_HOOKS" > "${CURSOR_HOOKS}.tmp" && mv "${CURSOR_HOOKS}.tmp" "$CURSOR_HOOKS"
    info "Added Cursor hooks to $CURSOR_HOOKS"
  fi
else
  mkdir -p .cursor
  jq -n --arg cmd "$HOOK_SCRIPT_PATH" '{
    version: 1,
    hooks: {
      preToolUse: [{
        command: $cmd,
        matcher: "Shell|Read"
      }]
    }
  }' > "$CURSOR_HOOKS"
  info "Created Cursor hooks at $CURSOR_HOOKS"
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
