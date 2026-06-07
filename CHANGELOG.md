# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `LICENSE` (MIT), `README.md`, `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, `SUPPORT.md`,
  and `SECURITY.md` community-health files.
- `.gitignore` for Go build outputs and release artifacts.
- `.github/dependabot.yml` for `gomod` and `github-actions` updates.
- `.golangci.yml` with a `depguard` stdlib-only allowlist and a security-focused linter
  set.
- CodeQL (`codeql.yml`) and supplementary security scanning (`security.yml`:
  `govulncheck`, `golangci-lint`/`gosec`, dependency review).
- Architecture Decision Records under `docs/decisions/` and a STRIDE threat model.
- Package documentation (`doc.go`) for `cmd/hashctl` and `internal/hashctl`.
- A **Memory & context learning** section in `.github/copilot-instructions.md`.
- GoReleaser release pipeline (`.goreleaser.yaml` + `release.yml`): tag-triggered,
  cross-platform archives with per-file checksums, and SLSA build-provenance attestation.
- OpenSSF Scorecard workflow (`scorecard.yml`) and a `copilot-setup-steps.yml` to bootstrap
  the Copilot cloud-agent environment.
- Copilot agent skills (`go-validation-gate`, `code-review`, `security-review`,
  `release-notes-writer`) and prompt files (`generate-go-tests`, `update-godoc`).
- A host-agnostic Go 1.25 development sandbox: `.devcontainer/` (devcontainer.json +
  Dockerfile) and `scripts/sandbox.sh`, which build and test the project inside a
  `golang:1.25` container (offline, stdlib-only, as the host user) so contributors and
  agents can build/test without installing the Go toolchain on the host.
- Populated the Copilot Memory Bank (`memory-bank/` core files: projectbrief,
  productContext, systemPatterns, techContext, activeContext, progress, copilot-rules)
  and integrated hashctl project instructions into `.github/copilot-instructions.md`
  alongside the Memory Bank framework.

### Changed

- Bumped the minimum Go version from `1.19` (end-of-life) to `1.25.0` (supported).
- Hardened `build-hashctl.yml`: least-privilege per-job permissions, SHA-pinned actions,
  and upgraded the gitleaks action to v3.
- Decoupled release publishing from `build-hashctl.yml` (now handled by GoReleaser in
  `release.yml`); its build job is now a cross-platform packaging smoke test.
- `install-hashctl.sh` and `verify-hashctl-release.sh` verify checksums in a
  format-agnostic way, compatible with GoReleaser's per-file `.sha256` sidecars.

### Fixed

- `build-hashctl.yml` now uploads the packaged per-platform binary (and its checksum) as a
  downloadable workflow artifact. The build job previously built and verified the tarball
  but discarded it, so no binary was retrievable from CI runs.

## [0.1.0] - 2026-06-02

### Added

- Initial release of `hashctl`, a stdlib-only Go CLI client for the Hash Engine REST API.
- Commands: `health`, `job` (create/get/list/cancel), `manifest` (build/get),
  `sign` (first/second), `smoke` (single-job/multi-job), plus `help` and `version`.
- Human and JSON output modes with secret redaction on all output paths.
- Bearer-token authentication from environment or a `chmod 600` token file; HTTPS
  enforced for non-loopback hosts.
- Semantic exit codes (success, usage, transport, API 4xx/5xx, poll timeout).
- Cross-platform release builds (`linux/amd64`, `darwin/arm64`) with SHA-256 checksums.
- Credential scanning via gitleaks in CI and as a pre-commit hook.

[Unreleased]: https://github.com/adamdost-0/hashctl/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/adamdost-0/hashctl/releases/tag/v0.1.0
