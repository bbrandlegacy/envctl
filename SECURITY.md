# Security Policy

If you discover a security issue in `envctl`, please report it privately and do not open a public discussion issue.

## Reporting

1. Use GitHub's private vulnerability reporting / Security Advisory flow for `bbrandlegacy/envctl` when available.
2. If the advisory flow is unavailable, open a minimal public issue that says only: "Private security report requested" and do not include exploit details or secret material.
3. Include a short reproduction path, impact summary, affected version/commit, and whether any raw secret values may have been exposed.

## Scope

- This project prioritizes local encryption and disclosure-safe output.
- Raw secret output is intentionally scoped to explicit `secrets get` usage.
- `envctl run` and MCP exec can expose secrets if the child process prints them; report bugs where ENVCTL itself exposes values through safe surfaces such as `list`, `context`, or `diff`.

## Disclosure timeline

We aim to acknowledge reports within 72 hours and provide remediation guidance for confirmed vulnerabilities as soon as possible.
