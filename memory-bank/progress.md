# Progress

> What works, what's left, current status, and known issues.

## What works

- **CLI:** `health`, `job` (create/get/list/cancel), `manifest` (build/get),
  `sign` (first/second), `smoke` (single-job/multi-job), `help`, `version`.
- **Output:** human and JSON modes; secret redaction on all paths.
- **Security in code:** literal-token rejection, `chmod 600` token files, HTTPS enforcement,
  redaction — all test-covered.
- **Validation gate:** test → vet → gofmt → build all pass; `golangci-lint` reports 0 issues.
- **Toolchain:** builds on supported `go 1.25.0` (replaced end-of-life 1.19).
- **CI:** build/test/lint, gitleaks, cross-platform packaging; new CodeQL, govulncheck,
  dependency-review, and Scorecard workflows (validated with `actionlint`).
- **Release:** GoReleaser config validated locally with `goreleaser check` and a full
  `--snapshot` build; artifacts remain compatible with the install/verify scripts.
- **Security assessment tooling:** `scripts/security-assessment.sh` (offline, stdlib-only
  evidence collector), the `security-assessment` skill, and its prompt drive a repeatable
  cross-vendor multi-model review synthesized into GitHub issues.

## What's left

- Enforce the `main` branch ruleset (required reviews, signed commits, linear history).
- Publish the first release (`v0.1.0`) after merge.
- Optional: SBOM attachment, fuzz tests, full Diátaxis `docs/` tree, broader test coverage
  (e.g. `smoke single-job`, `config.go` edge cases).

## Known issues / notes

- **Open security issues #3–#8 (multi-model assessment).** Most important: **#3 (Critical)** —
  success output and `manifest get` file writes bypass `redact()` (only the error path
  redacts), violating ADR-0003. Also #4 (redact heuristic gaps), #5 (literal `--bearer-token`
  accepted after a subcommand), #6 (`https→http` redirect can leak the bearer token).
- `.github/COMPENSATING_CONTROLS.md` is stale (branch protection is active); update or retire.
- `JobRecord.Raw` is currently unused in serialization paths.
- Tests for `smoke single-job` and some `config.go` paths are only covered indirectly.
