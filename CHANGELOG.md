# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Security

- Route **all** user-facing success output and the `manifest get --output-file`
  write through the `redact()` choke point, fixing an ADR-0003 redaction gap
  (issue #3) where a malicious or compromised API — or a manifest embedding SAS
  blob URLs — could leak credentials (SAS tokens, `AccountKey=`, JWTs) to stdout,
  CI logs, and files. JSON output is now scrubbed with a structure-preserving
  redactor that decodes each string literal before redacting, so the document
  stays valid JSON; human and direct-print output is buffered and redacted in a
  single write so no token can straddle a write boundary; manifest XML is
  redacted before `os.WriteFile` (mode `0o600` preserved); and the JSON error
  payload now redacts `error_code`, `route`, `job_id`, and `correlation_id`.

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
