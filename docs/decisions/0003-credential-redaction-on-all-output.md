---
status: accepted
date: 2026-06-07
decision-makers: [adamdost-0]
---

# Credential Redaction on All Output Paths

## Context and Problem Statement

`hashctl` handles bearer tokens (JWTs), Azure SAS query parameters, and `AccountKey=`
credentials. These values can appear in API responses, URLs, and error messages. How do we
prevent secrets from leaking into stdout, stderr, logs, or CI transcripts?

## Decision Drivers

- A leaked credential in CI logs is a high-severity incident.
- Redaction must be impossible to forget when adding a new output path.
- It must cover both structured (JSON) and human output, including errors.

## Considered Options

- **Central `redact()` applied to all output and error rendering.**
- **Ad-hoc masking** at each call site.
- **Rely on callers** to avoid printing secrets.

## Decision Outcome

Chosen option: **a central `redact()`** (in `internal/hashctl/output.go`) through which all
user-facing strings are routed. It strips bearer/authorization headers, Azure SAS query
parameters, `AccountKey=`, generic `secret=`/`password=` patterns, JWTs, and high-entropy
tokens.

### Consequences

- Good: a single, testable choke point for secret handling.
- Good: defense in depth — even unexpected secrets in API responses are scrubbed.
- Bad: every new output path must deliberately route through `writeError`/`redact`; this
  is enforced by code review, not the compiler.

### Confirmation

`output_test.go` exercises `redact()` against representative secrets. Code review must
verify that any new `fmt.Fprintf(stdout/stderr, …)` path goes through redaction. The
`.gitleaks.toml` allowlist covers only deterministic test fixtures.
