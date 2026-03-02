---
name: agent-secret
description: |
  Secure secret management using the OS keychain. Use this skill whenever you need to:
  - Set up environment variables or secrets for a project
  - Inject API keys, database passwords, or other credentials into .env files
  - Check if secrets are configured without exposing their values
  - List keys in .env files (without showing values)
  - List available secrets stored in the keychain
  - Configure secrets for automated workflows

  TRIGGER AUTOMATICALLY when the user:
  - Mentions adding, setting, updating, injecting, or configuring ANY value in .env files
  - Says "add key", "set key", "add api", "set api" in context of config/environment files
  - Mentions API keys for services (Google Maps, Stripe, OpenAI, Supabase, Firebase, AWS, etc.)
  - Mentions credentials, passwords, tokens, secrets, or sensitive configuration
  - Wants to configure environment variables or .env files
  - Asks about storing or managing secrets securely
  - References any service that typically requires API keys (maps, payments, auth, databases)
  - Says phrases like "add [service] to .env" or "set up [service] api key"
  - Wants to see what keys exist in a .env file
---

# Agent Secret - Secure Secret Management

This skill enables you to manage secrets securely using the `agent-secret` CLI tool. Secrets are stored in the OS's native keychain (macOS Keychain, Windows Credential Manager, Linux Secret Service) and never exposed in terminal output.

## Supported File Types

agent-secret inject command only works with files that have `.env` in their name:

- `.env` - Environment files
- `.env.local`, `.env.development`, `.env.production` - Environment-specific configs
- `my.env`, `config.env` - Custom .env files

The tool preserves file structure including comments, empty lines, and formatting.

## Core Principle

**Never expose secret values.** This tool is designed so that secret values are:
- Stored via hidden input prompt
- Only written to target files (never displayed)
- Verifiable for existence without revealing content

## Key Concepts

### Secret Name vs File Key Name

Understanding this distinction is critical:

- **Secret Name**: How the secret is stored in the keychain (e.g., `PROJECTX_SUPABASE_KEY`)
- **File Key Name**: How it appears in the target file (e.g., `SUPABASE_KEY`)

Use the mapping syntax `SECRET_NAME:FILE_KEY` to bridge these:

```bash
# Secret stored as PROJECTX_SUPABASE_KEY → appears as SUPABASE_KEY in file
agent-secret inject PROJECTX_SUPABASE_KEY:SUPABASE_KEY

# Secret stored as MYAPP_DB_PASSWORD → appears as DATABASE_URL in file
agent-secret inject MYAPP_DB_PASSWORD:DATABASE_URL

# Secret stored as PROJECTX_AWS_KEY → appears as AWS_ACCESS_KEY_ID in .env.local
agent-secret inject PROJECTX_AWS_KEY:AWS_ACCESS_KEY_ID -f .env.local
```

### Intelligent Secret Matching

**CRITICAL: Be smart about matching user requests to stored secrets.**

#### Understanding Project Prefixes

Secret names often follow the pattern: `PROJECTNAME_SERVICENAME_KEYTYPE`

Examples:
- `TRAVELER_GOOGLE_MAPS_API_KEY` → Project: TRAVELER, Service: GOOGLE_MAPS, Type: API_KEY
- `MYAPP_STRIPE_SECRET_KEY` → Project: MYAPP, Service: STRIPE, Type: SECRET_KEY
- `PROJECTX_DATABASE_URL` → Project: PROJECTX, Service: DATABASE, Type: URL

Common prefixes: `TRAVELER_`, `MYAPP_`, `PROJECTX_`, `APP_`, project-specific names followed by underscore.

#### Service Name Matching

When a user mentions a service, match it to secrets containing that service name (case-insensitive, ignore spaces/underscores):

| User says | Should match secrets containing |
|-----------|--------------------------------|
| "googlemaps", "google maps" | `GOOGLE_MAPS` |
| "stripe" | `STRIPE` |
| "supabase" | `SUPABASE` |
| "firebase" | `FIREBASE` |
| "database", "db" | `DATABASE`, `DB` |
| "aws" | `AWS` |
| "openai" | `OPENAI` |

#### Checking .env Files - Use FILE KEY, Not Secret Name

When checking if a key exists in .env, use the **FILE KEY NAME** (without project prefix), NOT the full secret name:

```bash
# WRONG - checking with full secret name
agent-secret check TRAVELER_GOOGLE_MAPS_API_KEY -f .env

# CORRECT - checking with file key name (without prefix)
agent-secret check GOOGLE_MAPS_API_KEY -f .env
```

#### Smart Discovery Workflow

When user asks to "add googlemaps" or "set up stripe":

1. **List secrets** to see what's available:
   ```bash
   agent-secret list
   ```

2. **Match the service** to stored secrets:
   - User says "googlemaps" → Look for secrets containing `GOOGLE_MAPS`
   - Found: `TRAVELER_GOOGLE_MAPS_API_KEY`

3. **Extract the file key name** (remove project prefix):
   - `TRAVELER_GOOGLE_MAPS_API_KEY` → File key: `GOOGLE_MAPS_API_KEY`

4. **Inject with mapping**:
   ```bash
   agent-secret inject TRAVELER_GOOGLE_MAPS_API_KEY:GOOGLE_MAPS_API_KEY -f .env
   ```

5. **Verify with file key name**:
   ```bash
   agent-secret check GOOGLE_MAPS_API_KEY -f .env
   ```

## Agent Decision Flow

When you need to work with secrets, follow this decision tree:

```
┌─────────────────────────────────────────────────────────────────┐
│ Do you know the SECRET_NAME?                                    │
├─────────────────────────────────────────────────────────────────┤
│ NO  → Run `agent-secret list` to discover available secrets     │
│     → If still unclear, ask user which secret to use            │
│                                                                 │
│ YES → Do you need to verify it exists?                          │
│     → Run `agent-secret check <SECRET_NAME> -q`                 │
│                                                                 │
│     → Do you need to list all keys in a file?                   │
│     → Run `agent-secret check --list -f <file>`                 │
│                                                                 │
│     → Do you need to create/update a key-value file?            │
│     → Run `agent-secret inject <SECRET_NAME>:<KEY_NAME>`        │
└─────────────────────────────────────────────────────────────────┘
```

### Smart Discovery Pattern

Before asking the user for a secret name, ALWAYS check what's available:

```bash
# First, discover existing secrets
agent-secret list

# Look for patterns like PROJECTNAME_SERVICENAME_KEY
# Then ask user to confirm if needed
```

## Commands Reference

### Store a Secret

```bash
agent-secret set <SECRET_NAME>
```

This prompts for the secret value with hidden input. Use this when a user needs to store a new credential.

Example:
```bash
agent-secret set API_KEY
# Enter value for API_KEY: **** (hidden)
# ✓ Secret 'API_KEY' stored successfully
```

### List Stored Secrets

```bash
agent-secret list
```

Shows the names of all stored secrets (not their values).

### Check if a Secret is Configured

```bash
agent-secret check <KEY_NAME> [-q]
```

Verifies a key exists in the .env file without showing the value.
- Without `-q`: Prints confirmation with value length
- With `-q` (quiet): Returns exit code only (0 = set, 1 = not set)

Useful for scripts and verification:
```bash
agent-secret check DATABASE_URL -q && echo "Configured" || echo "Missing"
```

### List Keys in .env File

```bash
agent-secret check --list [-f <file>]
agent-secret check -l [-f <file>]
```

Lists all keys in the target .env file without showing their values. Keys are sorted alphabetically.

Examples:
```bash
# List all keys in default ./.env
agent-secret check --list

# List keys from a specific file
agent-secret check --list -f ./config/.env
agent-secret check -l -f .env.local
```

This is useful for:
- Discovering what configuration a project needs
- Verifying which keys are already set up
- Comparing keys between different environments

### Inject Secrets into Key-Value Files

```bash
agent-secret inject <SECRET_SPEC>... [-f <file>]
```

Injects one or more secrets into a key-value file. Each argument can be:
- **Simple**: `SECRET_NAME` (uses same name as key)
- **Mapped**: `SECRET_NAME:KEY_NAME` (maps to different key name)

Examples:
```bash
# Single secret (default: .env)
agent-secret inject API_KEY

# Multiple secrets
agent-secret inject API_KEY DB_PASSWORD JWT_SECRET

# Map secret to different key name
agent-secret inject PROJECTX_JWT_SECRET:JWT_SECRET
agent-secret inject PROJECTX_DB_PASS:DB_PASSWORD

# Mixed mapping
agent-secret inject PROJECTX_JWT_SECRET:JWT_SECRET API_KEY DB_PASS:DATABASE_URL

# Environment-specific files (must contain '.env' in filename)
agent-secret inject API_KEY DB_URL -f .env.local
agent-secret inject API_KEY DB_URL -f .env.production
agent-secret inject API_KEY DB_URL -f ./config/.env
```

The injection preserves:
- Comments in the file
- Empty lines and whitespace
- Existing structure (updates in-place, appends new)

### Delete a Secret

```bash
agent-secret delete <SECRET_NAME> [--force]
```

Removes a secret from the keychain.
- Without `--force`: Prompts for confirmation
- With `--force`: Deletes immediately

## Workflow Patterns

### Pattern 1: Agent Needs to Create/Update Key-Value Files

When you need to set up credentials or configuration for a project:

1. **Discover available secrets** (if secret name not provided):
   ```bash
   agent-secret list
   ```

2. **Check if required secrets exist** before injecting:
   ```bash
   agent-secret check PROJECTX_SUPABASE_KEY -q && echo "Found" || echo "Missing"
   ```

3. **Inject with proper mapping** to get standard key names:
   ```bash
   # Environment files
   agent-secret inject PROJECTX_SUPABASE_KEY:SUPABASE_KEY PROJECTX_DB_URL:DATABASE_URL -f .env

   # Environment-specific
   agent-secret inject PROJECTX_API_KEY:API_KEY -f .env.production

   # Custom .env file locations
   agent-secret inject PROJECTX_AWS_KEY:AWS_ACCESS_KEY_ID -f ./config/.env
   ```

### Pattern 2: Agent Checking Prerequisites

Before running commands that require secrets, verify they're configured:

```bash
# Quiet mode returns exit code only (0 = set, 1 = not set)
agent-secret check SUPABASE_KEY -q || echo "Missing SUPABASE_KEY"
agent-secret check DATABASE_URL -q || echo "Missing DATABASE_URL"
```

If a secret is missing, inform the user they need to store it first:
```bash
agent-secret set PROJECTX_SUPABASE_KEY
```

### Pattern 3: User Provides Secret Name

When documentation or user specifies the secret name (e.g., "use PROJECTX_SUPABASE_KEY"):

```bash
# Verify it exists
agent-secret check PROJECTX_SUPABASE_KEY -q

# Inject with mapping to standard env key
agent-secret inject PROJECTX_SUPABASE_KEY:SUPABASE_KEY -f .env
```

### Pattern 4: User Mentions Service but Not Secret Name

User says: "I need to set up Supabase for ProjectX"

1. **List secrets first** to find matching pattern:
   ```bash
   agent-secret list
   # Look for: PROJECTX_SUPABASE_*, PROJECTX_SUPABASE_KEY, etc.
   ```

2. **If found**, confirm with user and inject to appropriate file:
   ```bash
   # Determine target file based on project structure
   agent-secret inject PROJECTX_SUPABASE_KEY:SUPABASE_KEY PROJECTX_SUPABASE_URL:SUPABASE_URL -f .env.local
   ```

3. **If not found**, ask user to store the secret first:
   ```bash
   agent-secret set PROJECTX_SUPABASE_KEY
   agent-secret set PROJECTX_SUPABASE_URL
   ```

### Setting Up a New Project

1. First, have the user store their secrets:
   ```bash
   agent-secret set PROJECT_API_KEY
   agent-secret set PROJECT_DB_PASSWORD
   ```

2. Then inject into the project's .env file:
   ```bash
   agent-secret inject PROJECT_API_KEY:API_KEY PROJECT_DB_PASSWORD:DB_PASSWORD -f ./project/.env
   ```

### Verifying Configuration Before Deployment

```bash
# Check all required secrets are configured
agent-secret check API_KEY -q || echo "Missing API_KEY"
agent-secret check DB_PASSWORD -q || echo "Missing DB_PASSWORD"
agent-secret check JWT_SECRET -q || echo "Missing JWT_SECRET"
```

### Multi-Project Setup

For users managing multiple projects with different credentials:

```bash
# Project X - environment files
agent-secret inject PROJECTX_API_KEY:API_KEY PROJECTX_DB:DATABASE_URL -f ./projectx/.env

# Project Y - environment-specific
agent-secret inject PROJECTY_API_KEY:API_KEY PROJECTY_DB:DATABASE_URL -f ./projecty/.env.production

# Project Z - custom location
agent-secret inject PROJECTZ_AWS_KEY:AWS_ACCESS_KEY_ID -f ./projectz/.env.local
```

## Security Best Practices

1. **Never try to read secret values** - They are intentionally hidden for security
2. **Use descriptive secret names** - Include project prefixes to avoid collisions
3. **Verify before injecting** - Use `check` to ensure secrets exist
4. **Use quiet mode in scripts** - `-q` flag for programmatic checks
5. **Map secrets to standard keys** - Use `SECRET_NAME:KEY_NAME` format to keep config files consistent

## Error Handling

- If a secret doesn't exist, `inject` will fail and list missing secrets
- `check` returns exit code 1 if the key is missing or empty
- All errors are written to stderr with helpful messages

## Agent Operating Rules

When working autonomously with secrets:

1. **Always use `list` before asking user** - Discover available secrets first
2. **Match services intelligently** - When user says "googlemaps", look for secrets containing `GOOGLE_MAPS`
3. **Understand project prefixes** - `TRAVELER_GOOGLE_MAPS_API_KEY` has prefix `TRAVELER_`, service `GOOGLE_MAPS`
4. **Check .env with FILE KEY names** - Use `GOOGLE_MAPS_API_KEY`, not `TRAVELER_GOOGLE_MAPS_API_KEY`
5. **Inject with mapping** - Always map `SECRET_NAME:FILE_KEY` to remove project prefix
6. **Use `check -q` to verify prerequisites** - Silent verification for scripts
7. **Never manually write credential files** - Always use `inject`
8. **Never attempt to read or log secret values** - Report status only
9. **Report configuration status, not values** - Say "configured" or "missing", never show values
10. **Only use files with '.env' in the name** - .env, .env.local, .env.production, etc.

## Common Scenarios

### Scenario: User says "add googlemaps" or "add google maps api key"

```bash
# 1. List secrets to find matching service
agent-secret list
# Output shows: TRAVELER_GOOGLE_MAPS_API_KEY

# 2. Extract file key name (remove project prefix TRAVELER_)
# File key = GOOGLE_MAPS_API_KEY

# 3. Inject with mapping
agent-secret inject TRAVELER_GOOGLE_MAPS_API_KEY:GOOGLE_MAPS_API_KEY -f .env

# 4. Verify using FILE KEY name (not the secret name)
agent-secret check GOOGLE_MAPS_API_KEY -f .env
```

### Scenario: Setting up configuration for a service

```bash
# User mentions: "Set up Supabase for this project"

# 1. Discover what secrets exist
agent-secret list
# Output shows: MYPROJECT_SUPABASE_KEY, MYPROJECT_SUPABASE_URL

# 2. Extract file key names (remove MYPROJECT_ prefix)
# File keys = SUPABASE_KEY, SUPABASE_URL

# 3. Inject with mapping
agent-secret inject MYPROJECT_SUPABASE_KEY:SUPABASE_KEY MYPROJECT_SUPABASE_URL:SUPABASE_URL -f .env.local

# 4. Verify using FILE KEY names
agent-secret check SUPABASE_KEY -f .env.local && \
agent-secret check SUPABASE_URL -f .env.local && \
echo "Supabase configured" || echo "Missing Supabase config"
```

### Scenario: Checking if project is configured

```bash
# Check using FILE KEY names (without project prefix)
agent-secret check DATABASE_URL -q && \
agent-secret check API_KEY -q && \
agent-secret check JWT_SECRET -q && \
echo "All secrets configured" || echo "Some secrets missing"
```

### Scenario: Missing secret handling

If `check` or `inject` fails due to missing secret:

1. Inform user: "Secret `PROJECTX_API_KEY` is not stored"
2. Provide command: "Run `agent-secret set PROJECTX_API_KEY` to store it"
3. Wait for user to complete storage before continuing

### Scenario: Discovering what keys a project needs

When you need to understand what configuration a project requires:

```bash
# List all keys in the .env file
agent-secret check --list -f .env
# Output:
# API_KEY
# DATABASE_URL
# JWT_SECRET
# REDIS_URL

# Then verify each is configured
agent-secret check API_KEY -q && echo "✓ API_KEY" || echo "✗ API_KEY"
agent-secret check DATABASE_URL -q && echo "✓ DATABASE_URL" || echo "✗ DATABASE_URL"
```

### Scenario: Comparing environments

When checking differences between development and production:

```bash
# List keys in each environment
agent-secret check --list -f .env.development
agent-secret check --list -f .env.production

# Compare outputs to ensure parity
```
