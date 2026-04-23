# agent-secret

A secure, developer-friendly CLI tool that acts as a **Local Secret Vault**. It eliminates the need to store sensitive credentials (API keys, database passwords) in plain-text `.env` files.

## Features

- **Secure Storage**: Secrets are stored in your OS's native keychain (macOS Keychain, Windows Credential Manager, Linux Secret Service)
- **No Secret Exposure**: Secrets are never displayed in the terminal - they are only written to target files
- **Smart Injection**: Injects secrets into `.env` files while preserving comments and formatting
- **Verification**: Check if secrets are configured without seeing their values
- **Hidden Input**: Secret values are never shown in terminal or command history
- **Cross-Platform**: Works on macOS, Windows, and Linux

## Installation

### Quick Install (recommended)

Installs the CLI, sets up hooks to protect your `.env` files from AI agents (Claude Code and Codex), and configures everything automatically:

```bash
curl -fsSL https://raw.githubusercontent.com/onurkerem/agent-secret/main/install.sh | bash
```

After installation, install the AI agent skill:

```bash
npx skills add https://github.com/onurkerem/agent-secret --skill agent-secret
```

### Homebrew

```bash
brew tap onurkerem/agent-secret
brew install agent-secret
```

### From Source

```bash
git clone https://github.com/onurkerem/agent-secret.git
cd agent-secret
go build -o agent-secret .
```

## Usage

### Store a secret

```bash
agent-secret set API_KEY
# Enter value for API_KEY: ****
# ✓ Secret 'API_KEY' stored successfully
```

### List stored secrets

```bash
agent-secret list
# Stored secrets (2):
#   - API_KEY
#   - DB_PASSWORD
```

### Check if a secret is configured

Verify that a key has a non-empty value in your .env file without seeing the actual value:

```bash
agent-secret check DATABASE_URL
# ✓ Key 'DATABASE_URL' is set (length: 32 characters)

# Quiet mode - use exit code only (useful for scripts)
agent-secret check API_KEY -q && echo "Configured" || echo "Missing"
```

### List keys in .env file

List all keys (not values) in a .env file:

```bash
# List all keys in default ./.env
agent-secret check --list
# API_KEY
# DATABASE_URL
# REDIS_URL

# List keys from a specific file
agent-secret check --list -f ./config/.env

# Short form
agent-secret check -l
```

### Inject secrets into .env file

The `inject` command only works with files that have `.env` in their name (e.g., `.env`, `.env.local`, `.env.production`).

```bash
# Inject single secret (secret name = key name)
agent-secret inject API_KEY

# Inject multiple secrets
agent-secret inject API_KEY DB_PASSWORD JWT_SECRET

# Map secret name to different key name (SECRET_NAME:KEY_NAME)
agent-secret inject PROJECTX_JWT_SECRET:JWT_SECRET
agent-secret inject PROJECTX_DB_PASS:DB_PASSWORD

# Mixed: some mapped, some not
agent-secret inject PROJECTX_JWT_SECRET:JWT_SECRET API_KEY DB_PASS:DATABASE_URL

# Specify custom .env file path (must contain '.env' in filename)
agent-secret inject API_KEY -f ./config/.env
agent-secret inject API_KEY -f .env.local
agent-secret inject API_KEY -f .env.production
```

### Delete a secret

```bash
# With confirmation
agent-secret delete API_KEY

# Force delete without confirmation
agent-secret delete API_KEY --force
```

## Security Model

`agent-secret` does not implement its own encryption. It relies entirely on your Operating System's native security infrastructure:

- **macOS**: Keychain Access
- **Windows**: Windows Credential Manager
- **Linux**: Secret Service API (e.g., GNOME Keyring, KWallet)

**Important**: `agent-secret` never exposes secret values directly in the terminal. Secrets can only be:
- Stored (via hidden prompt)
- Injected into files
- Verified for existence (without seeing the value)

## .env Parser Features

The smart parser preserves your `.env` file structure:

- Comments (`#`) are preserved
- Empty lines and whitespace are maintained
- Existing values are updated in-place
- New secrets are appended at the end
- Special characters in values are automatically quoted

Example `.env` before:
```env
# Database configuration
DB_HOST=localhost
DB_PORT=5432
```

After `agent-secret inject DB_PASSWORD`:
```env
# Database configuration
DB_HOST=localhost
DB_PORT=5432
DB_PASSWORD="my p@ssword!"
```

## For Automation Agents

The `check` and `inject` commands are designed for use by automated agents that need to:
- Verify secrets are properly configured (`check`)
- Set up environment files (`inject`)

These commands never expose actual secret values, making them safe for automated workflows.

## License

MIT License - see [LICENSE](LICENSE) for details.
