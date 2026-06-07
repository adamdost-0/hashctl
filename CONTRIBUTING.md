# Contributing to hashctl

Thanks for your interest in improving `hashctl`. This document describes how to build,
test, and submit changes.

## Prerequisites

- **Go** — the toolchain version is pinned via `go-version-file: go.mod`. No other
  tooling is required to build or test.
- **No third-party dependencies.** `hashctl` is **standard-library only** — there is no
  `go.sum`, and all Go commands run **offline** (`GOPROXY=off GOSUMDB=off`). Do not add
  `require` entries to `go.mod`.

## Validation gate (run before every commit/PR)

All four steps must pass; they mirror the CI gates in
`.github/workflows/build-hashctl.yml`:

```bash
# 1. Test
./scripts/test-hashctl.sh                                   # CGO_ENABLED=0 GOPROXY=off GOSUMDB=off go test ./...
# 2. Vet
CGO_ENABLED=0 GOPROXY=off GOSUMDB=off go vet ./...
# 3. Format check (must print nothing; auto-fix with: gofmt -w .)
gofmt -l .
# 4. Build
HASHCTL_VERSION=dev ./scripts/build-hashctl.sh              # writes bin/hashctl (untracked)
```

Optional, if installed: `golangci-lint run` (config in `.golangci.yml`) and
`pre-commit run --all-files` (gitleaks credential scan).

## Commit messages

Use [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` a new capability → **minor** version bump
- `fix:` a bug fix → **patch** version bump
- `feat!:` or a `BREAKING CHANGE:` footer → **major** version bump
- `chore:`, `ci:`, `docs:`, `test:`, `refactor:` → no version bump

Sign your commits (`git config commit.gpgsign true`); `main` requires signed commits.

## Security expectations

- Route every user-facing string through `redact()` (see `internal/hashctl/output.go`).
- Never commit secrets or fixtures that look like live credentials. Deterministic test
  literals must be allowlisted in `.gitleaks.toml`.
- Tokens are read only from `HASH_ENGINE_BEARER_TOKEN` or a `chmod 600`
  `--bearer-token-file`; the literal `--bearer-token` flag is rejected by design.

Report vulnerabilities privately — see [SECURITY.md](SECURITY.md).

## Documentation & memory

- Update `CHANGELOG.md` (Keep a Changelog format) and relevant `docs/` pages with your
  change.
- This repository uses **persistent agent memory** for context learning. When you (or an
  AI agent) discover a durable, reusable fact — a verified command, an architectural
  insight, a gotcha — capture it so the next contributor benefits. See
  `.github/copilot-instructions.md`.

## Pull requests

1. Branch from `main`, make your change, and run the validation gate.
2. Open a PR; all required status checks must pass.
3. A code owner ([`.github/CODEOWNERS`](.github/CODEOWNERS)) must approve before merge.
4. Keep history linear (squash or rebase); do not force-push to `main`.
