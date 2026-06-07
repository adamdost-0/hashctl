# Progress

> What works, what's left, current status, and known issues.

## What works

- **CLI:** `health`, `job` (create/get/list/cancel), `manifest` (build/get),
  `sign` (first/second), `smoke` (single-job/multi-job), `help`, `version`.
- **Output:** human and JSON modes; secret redaction on all paths.
- **Security in code:** literal-token rejection, `chmod 600` token files, HTTPS enforcement,
  redaction — all test-covered.
- **Validation gate:** test → vet → gofmt → build all pass; `golangci-lint` reports 0 issues.
- **CI:** build/test/lint, gitleaks, cross-platform packaging; new CodeQL, govulncheck,
  dependency-review, and Scorecard workflows (validated with `actionlint`).
- **Release:** GoReleaser config validated locally with `goreleaser check` and a full
  `--snapshot` build; artifacts remain compatible with the install/verify scripts.

## What's left

- Bump `go.mod` off the end-of-life `go 1.19` (see `activeContext.md`).
- Enforce the `main` branch ruleset (required reviews, signed commits, linear history).
- Publish the first release (`v0.1.0`) after merge.
- Optional: SBOM attachment, fuzz tests, full Diátaxis `docs/` tree, broader test coverage
  (e.g. `smoke single-job`, `config.go` edge cases).

## Known issues / notes

- `.github/COMPENSATING_CONTROLS.md` is stale (branch protection is active); update or retire.
- `JobRecord.Raw` is currently unused in serialization paths.
- Tests for `smoke single-job` and some `config.go` paths are only covered indirectly.
