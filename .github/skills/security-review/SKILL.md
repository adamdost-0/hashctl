---
name: security-review
description: Perform a security audit or threat-model review of hashctl. Use when asked to security-review, threat-model, or check hashctl for vulnerabilities. Focuses on credential handling, output redaction, and transport security.
---

# hashctl Security Review

hashctl is a CLI that handles bearer tokens (JWTs), Azure SAS query parameters, and
storage `AccountKey=` values. The primary attack surface is **credential leakage** through
output, logs, or error messages. This skill complements the gitleaks credential scan
(which finds literal secrets) with semantic analysis (paths that should redact but do not).
See `docs/threat-model.md` for the STRIDE model.

## Audit steps

### 1. Output redaction coverage (ADR-0003)
Trace every path from an HTTP response to output:
- `client.go` → `do()` → error construction
- `output.go` → `writeHuman()`, `writeJSON()`, `writeError()`
- every `fmt.Fprintf(stdout/stderr, …)` in `commands.go` and `smoke.go`

For each path, confirm the string is routed through `redact()`. Confirm `redact()` still
covers: `eyJ` (JWT), `sig=`/`sv=`/`se=` (SAS params), `AccountKey=`, and high-entropy tokens.

### 2. Token-file permissions (ADR-0004)
Confirm the `--bearer-token-file` handler rejects files whose permissions are broader than
`0600`, and that the literal `--bearer-token` flag remains rejected.

### 3. Transport security (ADR-0004)
Confirm `resolveConfig` (config.go) rejects non-loopback `http://` and that
`applyHeaders` (client.go) refuses a bearer token over non-loopback HTTP. Loopback
(`localhost`, `127.0.0.1`, `::1`) over HTTP is allowed for dev/smoke. Check there is test
coverage for both acceptance and rejection.

### 4. Dependency & supply chain
Confirm `go.mod` has no `require` entries and no `go.sum` (ADR-0001). Confirm CI runs
`govulncheck`, `gosec` (via golangci-lint), CodeQL, and dependency-review.

### 5. Test-fixture safety
No live credentials in fixtures; deterministic test literals are allowlisted only in
`.gitleaks.toml`. Run `pre-commit run --all-files` if available.

## Output

Report findings with file:line citations and a severity, mapped to the relevant STRIDE
category in `docs/threat-model.md`. Record any newly discovered durable invariant with
`store_memory`.
