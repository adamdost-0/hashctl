# System Patterns

> Architecture, key technical decisions, and component relationships. See `docs/decisions/`
> for the full Architecture Decision Records (ADRs).

## Package structure

- `cmd/hashctl/main.go` — an 11-line entrypoint: `os.Exit(hashctl.Run(args, stdout, stderr))`.
- `internal/hashctl/` — all logic (ADR-0006, enforced by Go's `internal/` rule):
  - `commands.go` — `Run`, the hand-rolled `parseGlobal`, and command dispatch.
  - `client.go` — `Client.do()` centralizes HTTP; typed errors `APIError`, `TransportError`, `PollError`.
  - `config.go` — `resolveConfig` precedence + URL/HTTPS validation.
  - `types.go` — exit-code constants and request/response structs.
  - `output.go` — human/JSON rendering, `redact()`, and `errorExitCode`.
  - `smoke.go` — end-to-end smoke flows; `help.go` — static help text.

## Key patterns

- **Dependency-injection seam:** the `app` struct (`stdout`, `stderr`, `getenv`,
  `httpClient`) is constructed via `New(...)` in tests with a fake env and an `httptest`
  client; `Run` is the integration boundary.
- **Single HTTP choke point:** every request goes through `Client.do()` →
  `applyHeaders()` (correlation id, local-auth headers, bearer + HTTPS guard).
- **Central redaction (ADR-0003):** all user-facing strings pass through `redact()`.
- **Semantic exit codes (ADR-0002):** mapped centrally by `errorExitCode`.
- **Config precedence (ADR-0005):** defaults → config file → env → CLI flags.
- **Two output modes only (ADR-0007):** `writeHuman` type-switches on the result; `writeJSON` otherwise.

## Architecture decisions (index)

ADR-0001 stdlib-only · 0002 semantic exit codes · 0003 redaction · 0004 HTTPS enforcement ·
0005 config precedence · 0006 internal-package boundary · 0007 dual output modes.

## Architecture governance (fitness functions)

- Go compiler enforces the `internal/` boundary.
- `.golangci.yml` `depguard` (list-mode strict; allow `$gostd` + this module) enforces stdlib-only.
- CI gates: test/vet/gofmt/build, golangci-lint, gosec, govulncheck, CodeQL, gitleaks.
