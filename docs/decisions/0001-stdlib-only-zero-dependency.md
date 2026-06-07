---
status: accepted
date: 2026-06-07
decision-makers: [adamdost-0]
---

# Stdlib-only, Zero External Dependency Policy

## Context and Problem Statement

`hashctl` is a CLI client for the Hash Engine REST API that participates in a
cryptographic signing workflow. The Go standard library already provides everything the
client needs: `net/http`, `encoding/json`, `net/url`, `flag`, `os`, and `time`. Should
`hashctl` be allowed to import third-party Go modules?

## Decision Drivers

- Minimize supply-chain attack surface in a security-sensitive signing workflow.
- Support fully offline / air-gapped builds (`GOPROXY=off GOSUMDB=off`).
- Keep security audits simple: reviewers need to understand one codebase only.
- Make the policy machine-verifiable: a `go.mod` with no `require` block and no `go.sum`.

## Considered Options

- **Stdlib only** — no `require` entries in `go.mod`.
- **Curated dependencies** — allow a small vetted set (e.g. a CLI framework).
- **Vendored dependencies** — allow modules but vendor them for reproducibility.

## Decision Outcome

Chosen option: **Stdlib only**, because it satisfies every decision driver and the
standard library is sufficient for the required functionality.

### Consequences

- Good: the supply-chain attack surface for this component is effectively zero.
- Good: CI runs `GOPROXY=off GOSUMDB=off` with no network access.
- Good: no `go.sum` means no dependency hash-mismatch incidents.
- Bad: CLI flag parsing (`parseGlobal`) is hand-rolled rather than using a framework and
  must be maintained as the CLI grows.

### Confirmation

1. **Compile-time:** `go.mod` has no `require` entries; any third-party import fails
   `go build` with `GOPROXY=off`.
2. **CI:** every Go command in `build-hashctl.yml` runs with `GOPROXY=off GOSUMDB=off`.
3. **Lint:** the `depguard` rule in `.golangci.yml` uses `list-mode: strict` with an
   allowlist of `$gostd` plus this module, so any non-stdlib import fails lint.
