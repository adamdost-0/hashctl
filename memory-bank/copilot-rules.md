# Copilot Rules

> The team's and Copilot's learning journal: implementation paths, preferences, security
> requirements, and evolving project patterns. Update this file when new rules emerge.

## 🚨 Security: Never commit secrets

- Never copy, move, or commit secret files or values (`.env`, `secrets.json`, API keys,
  tokens, passwords) to version control or into example/sample config files.
- Example files (e.g. `.env.example`) must contain only safe placeholder values.
- Verify no secrets are present before staging, committing, or pushing.
- If a secret is ever committed, treat it as an incident: remove it from history and rotate
  the affected credential immediately.
- Deterministic secret-like test literals must be allowlisted in `.gitleaks.toml` only.

## 🚨 Security: hashctl credential handling

- The literal `--bearer-token` flag is rejected by design. Tokens come only from
  `HASH_ENGINE_BEARER_TOKEN` or a `--bearer-token-file` verified as `chmod 600`.
- Never send a bearer token over non-loopback plaintext `http`; non-loopback API URLs must
  be `https`.
- Route **every** user-facing string (output and errors) through `redact()`.

## Engineering rules

- **Stdlib-only, offline:** no `require` entries in `go.mod`, no `go.sum`; every Go command
  runs with `GOPROXY=off GOSUMDB=off` and `CGO_ENABLED=0`.
- **Validation gate before every commit:** test → vet → gofmt → build (plus
  `golangci-lint run` if available). All must pass.
- **Conventional Commits + SemVer:** `feat:`/`fix:`/`feat!:`; update `CHANGELOG.md`; cut
  releases from a signed annotated tag `vX.Y.Z` matching `VERSION`.
- **Architecture changes need an ADR** in `docs/decisions/`; exit codes and the output
  contract are public — changing them is a breaking change.
- **Tests** use `net/http/httptest` only; no external frameworks.

## Memory & context learning

- Read all `/memory-bank/` files at the start of every task (see the checklist in
  `.github/copilot-instructions.md`).
- Update `activeContext.md` and `progress.md` after significant changes, and this file when
  a new rule or pattern is discovered.
- In addition to the file-based Memory Bank, capture durable, reusable facts with
  `store_memory` so they persist across sessions and repositories.

## Tool usage patterns

- Verify external resources before pinning (e.g. confirm GitHub Action commit SHAs with
  `git ls-remote` before committing a pin).
- Validate workflow YAML with `actionlint` and release config with `goreleaser check`
  before pushing.
