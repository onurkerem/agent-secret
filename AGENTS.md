# AGENTS.md

Read @README.md at every session start — it documents the CLI interface, security model, and user-facing behavior. This file covers the internal architecture, development patterns, and safety rules needed to work on the codebase.

## Project Overview

Go CLI tool (`go 1.21`) that stores secrets in the OS keychain and injects them into `.env` files. Released via GoReleaser with Homebrew distribution.

**Dependencies**: `spf13/cobra` (CLI framework), `zalando/go-keyring` (OS keychain), `golang.org/x/term` (hidden terminal input).

## Architecture

```
main.go              → calls cmd.Execute()
cmd/                 → Cobra commands, each file self-registers via init()
  root.go            → constants: serviceName="agent-secret", version
  set.go / inject.go / check.go / list.go / delete.go
internal/
  keyring/           → thin wrapper over go-keyring (Set/Get/Delete/Exists/List)
  parser/            → .env file parser and writer
  prompt/            → raw terminal mode for hidden input
skills/agent-secret/ → SKILL.md for AI agent integration
```

Commands use `os.Exit()` directly — errors are printed to stderr and the process terminates. There is no error propagation to `main()`.

## Key Internal Patterns

### Keyring Index (`internal/keyring`)

`go-keyring` has no native "list all keys" operation. The code maintains a metadata entry under the key `__agent_secret_index__` — a newline-separated list of secret names. `SetWithIndex` and `DeleteWithIndex` keep this index in sync.

When adding new keyring operations, always use the `*WithIndex` variants for any mutation. `List()` reads from this index, not the keychain itself.

### .env Parser (`internal/parser`)

- `InjectSecrets` does a two-pass write: first updates existing keys in-place, then appends new keys at the end. This preserves comments, blank lines, and ordering.
- Values containing spaces or special chars are auto-wrapped in double quotes via `formatValue`.
- Written files use permission `0600` (owner read/write only).
- `ParseEnvFile` returns an empty map (not error) for missing files.
- `ValidateEnvKey` enforces `[a-zA-Z_][a-zA-Z0-9_]*` via compiled regexp.

### Command Registration

Each command file defines a `var xxxCmd = &cobra.Command{...}` and registers it in `init()` via `rootCmd.AddCommand(xxxCmd)`. Flags are declared as package-level `var` and bound in `init()`.

## Development Rules

### Security — Mandatory

1. **Never print secret values.** All output must show only key names, lengths, or existence status. If you add a new command or flag, ensure it cannot leak values.
2. **Never read `.env` file contents for debugging.** If you need to verify injection worked, use `agent-secret check` or check file length/existence — do not cat or log file contents.
3. **Always use `.env` in filenames for example files.** Never create a file literally named `.env` in documentation or tests — use names like `env.example` or `test.env`.
4. **The `inject` command validates `.env` is in the filename** before writing. Do not bypass this check.
5. **File writes must use permission `0600`.** Do not change to more permissive modes.

### Code Style

- No external test frameworks or mock libraries. Use standard `testing` + table-driven tests with `t.Run`.
- No custom error types. Use `fmt.Errorf("context: %w", err)` for wrapping.
- No logging framework. Use `fmt.Fprintf` to `os.Stdout` or `os.Stderr`.
- No comments on what code does — name things clearly instead. Add comments only for non-obvious constraints or workarounds.
- No config files or environment variables for tool behavior. Constants are in `cmd/root.go`.

### Testing

- Only `internal/parser` has tests. The keyring and prompt packages interact with OS services and are not mocked.
- Use `t.TempDir()` for file I/O tests (auto-cleaned up).
- CI runs `go test -v -race ./...` — all tests must be race-free.
- `golangci-lint` runs in CI — fix lint errors before pushing.

### Adding New Commands

1. Create `cmd/<command>.go` with a `var <command>Cmd` and `init()` that calls `rootCmd.AddCommand`.
2. Use the `serviceName` constant from `root.go` for all keychain operations.
3. Print errors to `os.Stderr`, call `os.Exit(1)` on failure.
4. Never expose secret values in output.
5. Add flag bindings in `init()`, using package-level vars.

## Build & Release

- No Makefile. Use `go build -o agent-secret .` for local builds.
- GoReleaser handles cross-compilation (linux/darwin/windows, amd64/arm64) with `CGO_ENABLED=0`.
- Version is hardcoded in `cmd/root.go`. GoReleaser ldflags attempt to inject `main.version`/`commit`/`date` but the `main` package doesn't declare these vars — this is a known gap.
- Homebrew tap: `onurkerem/homebrew-agent-secret`.
- Release triggers on `v*` tags pushed to GitHub.

## Cross-Platform Notes

- `go-keyring` uses CGO on macOS. With `CGO_ENABLED=0` in GoReleaser, the static build may have issues on Linux (D-Bus requires a running session). `IsKeyringError` detects headless environments.
- `prompt.go` handles backspace (ASCII 127 and 8), Ctrl+C (ASCII 3), and only accepts printable chars (ASCII 32-126). Do not add platform-specific terminal handling.
