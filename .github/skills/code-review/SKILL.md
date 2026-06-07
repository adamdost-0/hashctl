---
name: code-review
description: Review hashctl pull requests and diffs against project-specific invariants. Use when reviewing a PR, reviewing a diff, or asked for a code review in this repository. Checks security invariants, exit-code semantics, the stdlib-only policy, and Go conventions.
---

# hashctl Code Review Checklist

Apply this checklist to every change. Cite file and line for each finding.

## Security invariants (check every PR)

1. **Redaction coverage.** Any new string that reaches user-facing output
   (`writeHuman`, `writeJSON`, or any `fmt.Fprintf(stdout/stderr, …)`) MUST be routed
   through `redact()`. Look for JWTs, SAS query params, `AccountKey=`, and high-entropy
   tokens that could leak (ADR-0003).
2. **Bearer-token handling.** The literal `--bearer-token` flag must stay rejected. Tokens
   come only from `HASH_ENGINE_BEARER_TOKEN` or a `--bearer-token-file` verified as
   `chmod 600`.
3. **HTTPS enforcement.** Non-loopback API URLs must be rejected unless `https` (ADR-0004),
   at both config time (`config.go`) and request time (`client.go applyHeaders`).
4. **No third-party imports.** `go.mod` must have zero `require` entries and there must be
   no `go.sum` (ADR-0001). A new import will also fail the `depguard` lint rule.

## Exit-code semantics (ADR-0002, types.go)

Verify the codes are unchanged: 0 success, 2 usage, 3 transport, 4 API 4xx, 5 API 5xx,
6 poll timeout. New error paths must map through `errorExitCode()`. Changing a code is a
MAJOR (breaking) change.

## Go conventions

- Global flags must precede the command (`parseGlobal` stops at the first non-flag arg).
- A new result type requires a new `case` in the `writeHuman` type switch (ADR-0007).
- Tests use `net/http/httptest` servers only — no external frameworks.
- Code must pass `gofmt`, `go vet`, and `golangci-lint run` (see the go-validation-gate skill).

## Documentation

- User-facing changes update `CHANGELOG.md` (`[Unreleased]`) and relevant docs.
- A new architectural decision (or a change to an existing one) needs an ADR in
  `docs/decisions/`.
