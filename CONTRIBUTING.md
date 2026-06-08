# Contributing to envctl

Thanks for helping improve `envctl`.

## Core contribution rules

- Keep changes focused on value-safety and explicit behavior.
- Avoid adding output paths that expose secret values.
- Preserve profile-scoped behavior and deterministic output ordering.
- Keep CLI and execution behavior backward compatible unless the story explicitly allows change.

## Development setup

- Install Go 1.22+
- Run `go build ./...` and manual smoke checks.
- Use `./envctl --help` to verify command contracts after changes.

## Commit style

Use concise, scoped messages. Prefer conventional commits:

- `feat: ...`
- `fix: ...`
- `docs: ...`
- `chore: ...`

## Testing guidance

- Add/adjust unit tests for touched logic when practical.
- Add/adjust fixture coverage for `.envdesc`, context output, and command execution behavior.
