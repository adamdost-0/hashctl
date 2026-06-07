# Tech Context

> Technologies, development setup, constraints, and dependencies.

## Language & dependencies

- **Go**, version pinned by `go.mod` (currently `go 1.19`; a bump to a supported release is
  a pending decision — see `activeContext.md`). CI resolves it via `go-version-file: go.mod`.
- **Zero third-party dependencies** — standard library only, no `go.sum` (ADR-0001).
- All Go commands run **offline**: `GOPROXY=off GOSUMDB=off`, `CGO_ENABLED=0`.

## Validation gate (run before every commit/PR)

```bash
./scripts/test-hashctl.sh                                   # go test ./...
CGO_ENABLED=0 GOPROXY=off GOSUMDB=off go vet ./...
gofmt -l .                                                  # must be empty; fix: gofmt -w .
HASHCTL_VERSION=dev ./scripts/build-hashctl.sh             # writes bin/hashctl
```

Optional: `golangci-lint run ./...` (config `.golangci.yml`). Single test:
`go test ./internal/hashctl -run TestName -v`.

## Scripts

- `build-hashctl.sh` (ldflags version injection) · `test-hashctl.sh` · `package-hashctl.sh`
  (tar.gz + sha256) · `get-hashctl-version.sh` · `install-hashctl.sh` ·
  `verify-hashctl-release.sh` (checksum verification is format-agnostic, GoReleaser-compatible).

## CI/CD

- `build-hashctl.yml` — test/vet/gofmt/build, gitleaks scan, cross-platform packaging.
- `security.yml` — golangci-lint, govulncheck, dependency-review.
- `codeql.yml` — CodeQL SAST. `scorecard.yml` — OpenSSF Scorecard.
- `release.yml` — GoReleaser (tag-triggered) with SLSA build-provenance attestation.
- `dependabot.yml` — gomod + github-actions. All actions are SHA-pinned.

## Constraints

- Never add `require` entries to `go.mod` or introduce a `go.sum`.
- Never set `GOPROXY` to anything other than `off`.
- Releases must come from an annotated, signed tag `vX.Y.Z`; `VERSION` must match the tag.
