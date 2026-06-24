# envctl

A secure, profile-aware CLI for managing environment variables with local encryption and AI-safe workflows.

`envctl` keeps secrets encrypted at rest using [age](https://github.com/FiloSottile/age), injects values into child processes only when explicitly running commands, and exposes context/list/diff views without leaking raw secret values.

## What envctl is for

- Local `.env` workflows that keep secrets local and encrypted.
- Developers who use multiple runtime environments (dev, staging, production) and want profile separation.
- AI-assisted teams that need strict separation between environment metadata and secret values.
- Local agent workflows that need safe env context and controlled runtime injection.

## Core features

- [x] Encrypted local vault (`.envctl/vault.age`) with passphrase-protected age encryption.
- [x] Atomic vault saves with restrictive local file permissions.
- [x] Profile lifecycle: create, list, switch, and delete.
- [x] Secret CRUD scoped by profile.
- [x] Safer secret entry via `--stdin` or `--prompt`.
- [x] Value-safe listing/context/diff views.
- [x] Command execution with runtime environment injection.
- [x] `.envdesc` metadata parser for variable typing and documentation.
- [x] Optional AI integrations:
  - Safe-by-default skill manifest generation (`envctl ai install-skill`).
  - Experimental MCP-style stdio adapter (`envctl mcp`) for tool-calling integrations.

## Requirements

- Go 1.22+
- macOS / Linux / Windows shell environment

## Project status

ENVCTL is a release-hardened local CLI baseline. Core CLI behavior, safe-output redaction, MCP exec gating, manifest safety modes, and vault persistence are covered by automated tests.

MCP support is intentionally described as experimental/MCP-style until full protocol compliance is audited against the current MCP specification.

## Install and run

### Build locally

```bash
git clone https://github.com/bbrandlegacy/envctl.git
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

Prefer `--stdin` or `--prompt` for real secrets so values do not land in shell history.

```bash
envctl profile create dev
envctl profile use dev

printf 'postgres://localhost:5432/dev\n' | envctl secrets set DATABASE_URL --stdin
envctl secrets set API_TOKEN --prompt

envctl secrets list
envctl context --profile dev --envdesc .envdesc

envctl run --profile dev -- your-command
```

For automation or non-sensitive examples, positional values are still supported:

```bash
envctl secrets set DEBUG true
```

## Command reference

```text
envctl init [--force]
envctl profile create <name>
envctl profile list
envctl profile use <name>
envctl profile delete <name> [--force]

envctl secrets set <KEY> <VALUE> [--profile <name>]
envctl secrets set <KEY> --stdin [--profile <name>]
envctl secrets set <KEY> --prompt [--profile <name>]
envctl secrets get <KEY> [--profile <name>]
envctl secrets unset <KEY> [--profile <name>]
envctl secrets list [--profile <name>] [--json]

envctl context [--profile <name>] [--envdesc .envdesc] [--json]
envctl diff <PROFILE_A> <PROFILE_B> [--json]
envctl run --profile <name> -- [command] [args...]

envctl ai install-skill [--target generic|claude|chatgpt|cursor|openai-functions] [--path PATH] [--global] [--apply]
envctl ai install-skill --include-exec [other flags]
envctl ai install-skill --include-sensitive-get [other flags]
envctl mcp [--transport stdio] [--allow-exec]
```

### Notes

- Use `--force` only when you explicitly want destructive overwrite behavior.
- `--` separates `envctl run` options from command arguments.
- `secrets get` intentionally prints a raw value. Treat it as sensitive output.
- `run` injects secrets into the child process. If the child prints its environment or config, the child can expose secrets.

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

Generate safe-by-default integration manifests for non-MCP clients:

```bash
envctl ai install-skill --target generic
```

The default manifest includes safe context only. It omits command execution and raw secret retrieval.

Write to platform-specific location:

```bash
envctl ai install-skill --target claude --apply
```

Use explicit path when needed:

```bash
envctl ai install-skill --target cursor --path ./envctl-cursor-skill.json --apply
```

Only for trusted local workflows that explicitly need command execution with injected secrets, include the exec command:

```bash
envctl ai install-skill --target generic --include-exec
```

Only for trusted local workflows that explicitly need raw value retrieval, include the sensitive `envctl_get` command:

```bash
envctl ai install-skill --target generic --include-sensitive-get
```

You may combine both privileged modes, but do so only for trusted local operators:

```bash
envctl ai install-skill --target generic --include-exec --include-sensitive-get
```

### MCP-style adapter (experimental)

Start a stdio JSON-RPC/MCP-style adapter:

```bash
envctl mcp --transport stdio
```

Supported methods:

- `initialize`
- `tools/list`
- `tools/call`
  - `name: envctl_context` — safe context; no raw values.
  - `name: envctl_exec` — command execution; disabled by default.

Enable exec only for trusted local clients:

```bash
envctl mcp --transport stdio --allow-exec
# or
ENVCTL_MCP_ALLOW_EXEC=1 envctl mcp --transport stdio
```

`envctl_exec` can return child stdout/stderr. Treat that output as sensitive if the child may print secrets.

## Security model

- Vault is encrypted using passphrase-protected `age` at rest.
- Vault saves use temp-file + rename replacement and restrictive permissions where supported.
- `list`, `context`, and `diff` never print raw secret values.
- `get` is intentionally the explicit command that prints a raw value.
- `run` and MCP exec inject values into child process environments; child output is outside ENVCTL's redaction guarantee.
- Prefer `secrets set --stdin` or `secrets set --prompt` for real secrets.
- Default passphrase source precedence:
  1. `--passphrase-file`
  2. `ENVCTL_PASSPHRASE`

## Error behavior (high-level)

- Missing or invalid vault: user-friendly error with setup guidance.
- Wrong passphrase: non-disclosure failures without exposing secret material.
- Command execution failures preserve child exit behavior where possible.

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
- `_bmad/`
- `.agents/`
- `_bmad-output/`
- local agent handoff/generated scan docs under `docs/`
- `.envctl/`
- `.envctl-pass`

## Development notes

If you want to extend `envctl`, start from the command and application layers under `internal/` and preserve the value-safety guarantees above.

Recommended validation:

```bash
go mod tidy
gofmt -w cmd internal
go test ./...
go vet ./...
go test -race ./...
go test -cover ./...
go build -o /tmp/envctl-release/envctl ./cmd/envctl
git diff --check
gofmt -l cmd internal
```

## Roadmap

- Optional TUI for profile browser workflows.
- Encrypted sharing/export workflows based on age recipients.
- Full MCP compliance audit and protocol hardening.

## License

This project is available under the MIT License.
