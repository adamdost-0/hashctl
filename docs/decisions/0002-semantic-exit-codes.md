---
status: accepted
date: 2026-06-07
decision-makers: [adamdost-0]
---

# Semantic Exit Codes

## Context and Problem Statement

`hashctl` is intended to be scripted in CI pipelines. Callers need to distinguish failure
classes (bad input, network failure, API error, timeout) without parsing output. What exit
code scheme should the CLI use?

## Decision Drivers

- Scripts must branch on failure class deterministically.
- Codes must be stable across releases (changing them is a breaking change).
- The mapping must be centralized so every error path is consistent.

## Considered Options

- **A small set of semantic codes** mapped from typed errors.
- **Binary success/failure** (0 / 1) only.
- **Mirror HTTP status codes** directly as exit codes.

## Decision Outcome

Chosen option: **a small set of semantic codes**, defined as constants and assigned by a
single mapping function. This is expressive without leaking HTTP specifics into the exit
contract.

The codes (see `internal/hashctl/types.go`) are:

| Code | Constant | Meaning |
|---|---|---|
| 0 | `ExitSuccess` | Success |
| 2 | `ExitUsage` | Usage/config error, unknown command |
| 3 | `ExitTransport` | Network/transport failure |
| 4 | `ExitAPIClient` | API returned HTTP 4xx |
| 5 | `ExitAPIServer` | API returned HTTP 5xx |
| 6 | `ExitPollTimeout` | Job did not reach the target state in time |

### Consequences

- Good: callers can branch on failure class reliably.
- Good: `errorExitCode` (in `output.go`) is the single source of truth.
- Bad: exit codes are now part of the public contract and changing them is a MAJOR change.

### Confirmation

`errorExitCode` is covered by `commands_test.go`; the table above must match the constants
in `internal/hashctl/types.go` and the `docs`/README exit-code table.
