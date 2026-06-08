# Product Brief: envctl

**Date:** 2026-06-08
**Author:** ocinbuona
**Version:** 1.0
**Project Type:** CLI Tool
**Project Level:** 2

---

## Executive Summary

`envctl` is a Go CLI tool for encrypted, profile-based `.env` file management with AI-safe secret injection. It targets developers who juggle multiple environments and need to keep secrets out of git, out of AI token streams, and out of copy-paste workflows. It fills the gap between heavyweight platforms (Infisical, Doppler) and zero-security plain `.env` files.

---

## Problem Statement

### The Problem

Developers constantly deal with `.env` files containing sensitive secrets (API keys, database URLs, tokens) that:
- Cannot be committed to git safely
- Must be manually transferred between machines/teammates
- Get pasted into AI assistants (ChatGPT, Claude) — leaking secrets into token streams and logs
- Are easy to lose, expose, or confuse across environments (dev/staging/prod)

No existing tool solves all three: local-first, encrypted at rest, AND AI-safe execution proxy.

### Why Now?

AI coding assistants (Claude, Cursor, Copilot) are now running shell commands on behalf of developers. When AI runs `psql $DATABASE_URL` or `curl -H "Authorization: $API_KEY"`, the secret either leaks into the prompt or the developer must manually inject it — breaking the workflow. This problem didn't exist at scale until 2024-2025.

### Impact if Unsolved

- Secrets leak into AI provider logs via prompt context
- Developers keep using insecure workarounds (pasting secrets, committing `.env` files)
- No audit trail of what secrets were used when
- Team secret sharing stays ad-hoc (Slack DMs, shared Google Docs)

---

## Target Audience

### Primary Users

Solo developers and small teams (1-10 people) who:
- Write code daily (Go, Python, Node, etc.)
- Use AI coding assistants (Claude Code, Cursor, Copilot)
- Manage multiple environments (dev, staging, prod)
- Are comfortable with CLI tools
- Care about security but don't want to run infrastructure

### Secondary Users

- DevOps engineers who want git-safe encrypted env files without SOPS complexity
- AI agent builders who need to give agents env access without secret exposure
- Open source maintainers who need to document env shape without exposing values

### User Needs

- Store secrets encrypted locally — no cloud required
- Run any CLI command with secrets injected at execution time, never in the prompt
- Show AI assistants the env *shape* (variable names, types, descriptions) without the values
- Switch profiles (dev/staging/prod) with one command
- Diff environments to catch missing or mismatched variables

---

## Solution Overview

### Proposed Solution

Single Go binary. No server. No cloud account required. Secrets encrypted at rest with `age` encryption (modern, no GPG complexity). Three interaction modes: direct CLI commands, TUI browser, MCP server for AI agent integration.

### Key Features

**Core (v0.1)**
- `envctl init` — create encrypted vault with `age` (passphrase or key file)
- `envctl set KEY value --profile dev` — add/update a secret
- `envctl get KEY --profile dev` — retrieve a secret
- `envctl list --profile dev` — masked output (show keys, hide values)
- `envctl run --profile dev -- <any command>` — exec wrapper, injects env vars at runtime

**AI Features (v0.2)**
- `envctl context --profile dev` — safe schema dump: variable names, types, descriptions, SET/MISSING status — zero values for sensitive vars
- `envctl mcp` — starts MCP server exposing `envctl_exec` and `envctl_context` tools so Claude Desktop / Cursor can run commands with injected secrets without values touching the token stream
- `.envdesc` sidecar — per-variable description file AI reads for context

**TUI (v0.3)**
- Browse profiles, add/edit/delete vars, diff two profiles
- Built with `tcell` (same stack as PlexLinker)

**Sync / Share (v1.0)**
- Encrypted git export (commit `.env.vault` safely)
- Export to 1Password / Bitwarden CLI
- Encrypted file bundle for team sharing (no server needed)

### Value Proposition

`envctl` is the only tool that acts as a **secure execution proxy for AI agents** — AI calls a tool, secrets are injected at exec time, values never appear in the token stream. Combined with local-first encryption and zero infrastructure requirement, it fills a gap no existing tool addresses.

---

## Business Objectives

### Goals

- Ship v0.1 (core) within 2 weeks of project start
- Ship v0.2 (AI features) within 4 weeks
- Reach 100 GitHub stars within 60 days of public launch
- Get picked up by at least one AI tooling community (Cursor Discord, Claude subreddit, r/golang)
- Generate portfolio signal strong enough to support job applications

### Success Metrics

- GitHub stars: 100 in 60 days, 500 in 6 months
- Downloads via `go install` or Homebrew tap: 500 in 3 months
- At least one HN Show HN front page appearance
- MCP server feature cited in at least one community post/blog not written by author

### Business Value

- Portfolio: demonstrates Go, encryption, MCP protocol, TUI, UX thinking — strong hiring signal
- Monetization path: free core (MIT), paid team features (encrypted sync, audit logs, web dashboard)
- Community: MCP server feature is novel enough to generate organic sharing in AI dev communities

---

## Scope

### In Scope (v0.1 - v0.3)

- Local encrypted vault using `age` encryption
- Profile management (create, switch, list, delete profiles)
- CRUD operations on secrets per profile
- `run` command — exec wrapper injecting env at runtime
- `context` command — AI-safe schema dump
- MCP server with `envctl_exec` and `envctl_context` tools
- `.envdesc` sidecar support
- TUI browser for profiles and variables
- Profile diff command
- Single binary, cross-platform (macOS, Linux, Windows)
- `go install` and Homebrew tap distribution

### Out of Scope (this version)

- Cloud sync or cloud-hosted vault
- Web dashboard / GUI
- Team permission management
- Secret rotation
- Dynamic secrets
- Kubernetes / Helm integration
- CI/CD pipeline native integration (GitHub Actions, etc.)
- Audit logs beyond local file

### Future Considerations

- Encrypted git commit workflow (`.env.vault` format)
- Export to 1Password / Bitwarden CLI
- Team encrypted bundle sharing
- Self-hosted web dashboard
- GitHub Actions secret injection
- VS Code extension

---

## Stakeholders

- **ocinbuona (Builder/Owner)** — High influence. Solo developer, building for portfolio + potential monetization.
- **Developer community (Users)** — High influence. Adoption drives value. Early feedback shapes v1.0.
- **AI tooling communities** — Medium influence. Claude/Cursor/Copilot communities are distribution channels.

---

## Constraints and Assumptions

### Constraints

- Solo developer — scope must be shippable in 4 weeks to v0.2
- No budget for infrastructure — must be local-first, no server required
- Go only — no polyglot complexity, single binary requirement is non-negotiable
- Must not require users to have GPG installed — `age` is the encryption choice

### Assumptions

- Users are comfortable with CLI tools
- Target users already use AI coding assistants
- MCP protocol (Model Context Protocol) remains stable — it is an open standard
- `age` encryption library (`filippo.io/age`) is stable and well-maintained
- Users have Go installed OR will use pre-built binaries from GitHub releases

---

## Success Criteria

- `envctl run --profile prod -- psql -c "SELECT 1"` works end-to-end with secrets injected, nothing in shell history
- MCP server passes secrets to a real Claude Desktop / Cursor session without values appearing in conversation
- `envctl context` output is safe to paste directly into any AI assistant
- TUI is navigable without reading docs
- Single `go install` command works on macOS, Linux, Windows
- At least 3 developers outside the author use it and report it working

---

## Timeline

### Target Launch

- v0.1 public: 2 weeks from project start
- v0.2 (AI features): 4 weeks from project start
- v0.3 (TUI): 6 weeks from project start
- v1.0 (sync/share): 3 months from project start

### Key Milestones

- Project scaffold + `age` vault working: Day 3
- `envctl run` working end-to-end: Day 7
- `envctl context` + MCP server working: Day 21
- TUI browser working: Day 35
- Homebrew tap published: Day 42
- Show HN post: Day 42-45

---

## Risks

- **Risk:** MCP server adoption is still early — AI assistants may not support it widely enough for users to care
  - **Likelihood:** Medium
  - **Mitigation:** Ship `envctl context` (copy-paste workflow) as fallback; MCP is bonus not core

- **Risk:** Infisical / Dotenvx adds exec proxy + MCP before launch
  - **Likelihood:** Low
  - **Mitigation:** Ship fast (4 weeks to v0.2); being early matters more than being only

- **Risk:** `age` encryption has a breaking API change
  - **Likelihood:** Low
  - **Mitigation:** Pin dependency version; `filippo.io/age` is stable and authored by Go security team member

- **Risk:** Solo developer burnout / scope creep delays launch
  - **Likelihood:** Medium
  - **Mitigation:** Hard scope cutoff at v0.2 for first public launch; TUI is v0.3 not launch blocker

---

## Competitive Landscape

| Tool | Type | Encrypted | Local-first | AI-safe exec | TUI | Go binary |
|------|------|-----------|-------------|--------------|-----|-----------|
| **envctl** | Profile manager + AI proxy | Yes (age) | Yes | Yes (MCP) | Yes | Yes |
| Infisical | Enterprise platform | Yes | No (cloud) | No | No | No |
| Dotenvx | .env encryption | Yes | Yes | No | No | No (JS) |
| SOPS | File encryption | Yes | Yes | No | No | Yes |
| envio | Profile manager | Yes (GPG) | Yes | No | No | No (Rust) |
| Doppler | SaaS platform | Yes | No | No | No | No |

**envctl is the only tool combining:** local-first + age encryption + AI-safe exec proxy + MCP server + TUI.
