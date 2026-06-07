---
status: accepted
date: 2026-06-07
decision-makers: [adamdost-0]
---

# All Logic in the `internal/hashctl` Package

## Context and Problem Statement

`hashctl` is a single binary. Where should application logic live, and how do we prevent
the package from becoming an importable public API surface we do not intend to support?

## Decision Drivers

- The CLI is a product, not a library; we do not want to commit to a public Go API.
- Testability: the logic must be exercisable without invoking `os.Exit`.
- A clear, enforceable boundary between the entry point and the logic.

## Considered Options

- **Thin `cmd/hashctl` entry point; all logic in `internal/hashctl`.**
- Put logic in a public `pkg/` package (commits us to API stability).
- Put everything in `package main` (hard to unit test, no seam).

## Decision Outcome

Chosen option: **a thin `cmd/hashctl/main.go` that calls
`internal/hashctl.Run(args, stdout, stderr)` and exits with the returned code; all logic
lives in `internal/hashctl`.** The `app` struct (`stdout`, `stderr`, `getenv`,
`httpClient`) is the dependency-injection seam, constructed via `New(...)` in tests.

### Consequences

- Good: Go's `internal/` rule makes the package un-importable by other modules — no
  accidental public API.
- Good: `Run`/`New` provide a clean integration seam for tests with fake env and HTTP.
- Bad: the only legitimate cross-package import is `cmd/hashctl` → `internal/hashctl`.

### Confirmation

The Go compiler enforces `internal/` visibility. The `depguard` allowlist additionally
restricts imports to `$gostd` and this module, and `cmd/hashctl` is an 11-line entry point
reviewed to import only `internal/hashctl`.
