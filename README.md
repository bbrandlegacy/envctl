# envctl

A secure, profile-aware CLI for managing environment variables with local encryption and AI-safe workflows.

`envctl` keeps secrets encrypted at rest using [age](https://github.com/FiloSottile/age), injects values into child processes only when explicitly running commands, and exposes context and diff views without leaking raw secret values.

## What envctl is for

- Local `.env` workflows that keep secrets local and encrypted.
- Developers who use multiple runtime environments (dev, staging, production) and want profile separation.
- AI-assisted teams that still require strict separation between metadata and secret values.

## Core features

- [x] Encrypted local vault (`.envctl/vault.age`) with passphrase protection.
- [x] Profile lifecycle: create, list, switch, and delete.
- [x] Secret CRUD scoped by profile.
- [x] Value-safe listing/context/diff views.
- [x] Command execution with runtime environment injection.
- [x] `.envdesc` metadata parser for variable typing and documentation.
- [x] Optional AI integrations:
  - Skill manifest generation (`envctl ai install-skill`).
  - MCP stdio adapter (`envctl mcp`) for tool-calling integrations.

## Requirements

- Go 1.22+
- macOS / Linux / Windows shell environment

## Project status

- MVP CLI and AI-adjacent integration are implemented.
- MCP and AI skill installer are opt-in features and can be disabled/avoided.
- This repository is now prepared as a public-facing project with non-runtime planning artifacts removed from distribution.

## Install and run

### Build locally

```bash
git clone <your-github-org>/envctl.git
cd envctl

go build -o envctl ./cmd/envctl
./envctl --help
```

### Install a binary locally

```bash
go install ./cmd/envctl
# then run envctl directly from your PATH if Go bin is configured
```

### Initialize a vault

```bash
export ENVCTL_PASSPHRASE='replace-with-strong-password'
envctl init
```

Or automation-safe:

```bash
printf 'replace-with-strong-password\n' > .envctl-pass
envctl init --passphrase-file .envctl-pass
```

## Quick start

```bash
envctl profile create dev
envctl profile use dev
envctl secrets set DATABASE_URL 'postgres://localhost:5432/dev'
envctl secrets set API_TOKEN 'abc123'
envctl secrets list

envctl context --profile dev --envdesc .envdesc

envctl run --profile dev -- env | grep -E '(DATABASE_URL|API_TOKEN)'
```

## Command reference

```text
envctl init [--force]
envctl profile create <name>
envctl profile list
envctl profile use <name>
envctl profile delete <name> [--force]

envctl secrets set <KEY> <VALUE> [--profile <name>]
envctl secrets get <KEY> [--profile <name>]
envctl secrets unset <KEY> [--profile <name>]
envctl secrets list [--profile <name>] [--json]

envctl context [--profile <name>] [--envdesc .envdesc] [--json]
envctl diff <PROFILE_A> <PROFILE_B> [--json]
envctl run -- [command] [args...]

envctl ai install-skill [--target generic|claude|chatgpt|cursor|openai-functions] [--path PATH] [--global|--local] [--apply]
envctl mcp [--transport stdio]
```

### Notes

- Use `--force` only when you explicitly want destructive overwrite behavior.
- `--` separates `envctl run` options from command arguments.

## `.envdesc` metadata

`envctl` supports an optional sidecar file (default `.envdesc`) used for safe context output.

Format:

```text
KEY: type - description
OPTIONAL_KEY?: type - description
```

Example:

```text
DATABASE_URL: url - Primary PostgreSQL connection string
API_TOKEN: token - Service bearer token
DEBUG: bool - Enables debug mode
FEATURE_FLAG?: bool - Optional feature toggle
```

## AI integrations

### Skill manifest installer

Generate integration manifests for non-MCP clients:

```bash
envctl ai install-skill --target generic
```

Write to platform-specific location:

```bash
envctl ai install-skill --target claude --apply
```

Use explicit path when needed:

```bash
envctl ai install-skill --target cursor --path ./envctl-cursor-skill.json --apply
```

### MCP adapter (experimental)

Start a stdio JSON-RPC style adapter:

```bash
envctl mcp --transport stdio
```

Supported methods:

- `initialize`
- `tools/list`
- `tools/call`
  - `name: envctl_context`
  - `name: envctl_exec`

## Security model

- Vault is encrypted using passphrase-protected `age` at rest.
- `list`, `context`, and `diff` never print raw secret values.
- `get` is intentionally the explicit command that prints a value.
- `run` injects values into the child process environment without command output leaking those values by default.
- Default passphrase source precedence:
  1. `--passphrase-file`
  2. `ENVCTL_PASSPHRASE`

## Error behavior (high-level)

- Missing or invalid vault: user-friendly error with setup guidance.
- Wrong passphrase: non-disclosure failures without exposing secret material.
- Command execution failures return captured process output and exit behavior.

## Repository layout

```text
envctl/
  cmd/envctl/
  internal/
    ai/
    app/
    cli/
    crypto/
    domain/
    envdesc/
    output/
    runner/
    store/
```

## Distribution hygiene

The following paths are excluded from public artifacts:

- `.bmad/`
- `.agents/`
- `_bmad-output/`
- `.envctl/`
- `.envctl-pass`

## Development notes

If you want to extend `envctl`, start from the command and application layers under `internal/` and preserve the value-safety guarantees above.

## Roadmap

- Harden cross-platform shell quoting behavior.
- Optional TUI for profile browser workflows.
- Encrypted sharing/export workflows.

## License

This project is available under the MIT License. Add your preferred license text in `LICENSE` if you plan to publish.
