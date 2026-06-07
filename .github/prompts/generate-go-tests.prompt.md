---
mode: agent
description: Generate Go tests for hashctl following the established runTest/httptest pattern.
---

# Generate hashctl tests

Generate unit/integration tests for `${input:target:Which function, command, or file should be tested?}`
in `internal/hashctl/`, following the existing testing patterns. Standard library only.

## Patterns to follow (from commands_test.go / client_test.go)

1. Use the `runTest(args, env, handler)` helper: it starts a `net/http/httptest` server,
   constructs the app via `New(stdout, stderr, fakeEnv, server.Client())`, and returns
   `(exitCode, stdout, stderr)`. Use `runLocal(args, env)` for cases that need no server
   (help, version, flag rejection).
2. Replace `os.Getenv` with a fake env map via `testEnv(map)`.
3. Assert on all three of: exit code (semantic codes in `types.go`), stdout, and stderr.
4. Prefer table-driven tests (`[]struct{ name string; … }`).

## Requirements

- No external test frameworks — only the standard `testing` package.
- Tests must pass `go vet ./...`, `gofmt -l .` (no output), and `golangci-lint run`.
- Cover: happy path, an API error (e.g. 401/403 → exit 4), a transport error, a usage
  error (missing required flag → exit 2), and at least one edge case.
- Never embed live credentials; if a fixture looks secret-like, allowlist it in
  `.gitleaks.toml`.

## Output

A complete test file or additions to the relevant `_test.go`, with the matching package
declaration. After writing, run the go-validation-gate skill to confirm everything passes.
