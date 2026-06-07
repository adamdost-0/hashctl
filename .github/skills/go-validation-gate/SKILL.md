---
name: go-validation-gate
description: Run the hashctl validation gate (test, vet, gofmt, build) before any commit or pull request. Use this skill whenever asked to validate, check, verify, or prepare hashctl code for a commit or PR.
---

# hashctl Validation Gate

Run this four-step sequence in order. **All four must pass.** Stop and fix at the first
failure before continuing. Every Go command runs offline (stdlib-only project).

```bash
# 1. Test
./scripts/test-hashctl.sh                                   # CGO_ENABLED=0 GOPROXY=off GOSUMDB=off go test ./...
# 2. Vet
CGO_ENABLED=0 GOPROXY=off GOSUMDB=off go vet ./...
# 3. Format check (must print nothing; auto-fix with: gofmt -w .)
gofmt -l .
# 4. Build
HASHCTL_VERSION=dev ./scripts/build-hashctl.sh             # writes bin/hashctl (untracked)
```

If `golangci-lint` is available, also run it (config in `.golangci.yml`):

```bash
golangci-lint run ./...
```

## On failure

- **Test failure:** read the error and fix the failing test or the code under test.
- **Vet failure:** fix the reported issue; never suppress it.
- **gofmt failure:** run `gofmt -w .`, then re-run step 3 to confirm it is clean.
- **Build failure:** fix compilation errors (check for missing types in `types.go`).
- **Lint failure:** fix the finding; a new third-party import will trip the `depguard`
  stdlib-only rule — remove it (do not add dependencies).

## Environment invariants

- NEVER add `require` entries to `go.mod` — hashctl is standard-library only.
- NEVER set `GOPROXY` to anything other than `off` — no network downloads.
- `CGO_ENABLED=0` is mandatory for cross-compilation.

## Memory

If you discover a durable, reusable fact while validating (a new command, a flaky path, an
environment quirk), capture it with `store_memory` per `.github/copilot-instructions.md`.
