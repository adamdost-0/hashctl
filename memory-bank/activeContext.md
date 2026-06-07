# Active Context

> Current work focus, recent changes, next steps, and active decisions.
> Keep this file and `progress.md` the most up to date.

## Current focus

SDLC modernization: bringing `hashctl` up to Microsoft-aligned SDL/SDLC standards
(version control, lifecycle/release management, documentation, security testing,
architecture-driven design, and governance).

## Recent changes (branch `sdlc-modernization`)

- **Community health & legal:** `LICENSE` (MIT), `README.md`, `CONTRIBUTING.md`,
  `SECURITY.md`, `CODE_OF_CONDUCT.md`, `SUPPORT.md`, `CHANGELOG.md`, `.gitignore`,
  issue/PR templates.
- **Security CI:** hardened `build-hashctl.yml` (least-privilege permissions, SHA-pinned
  actions, gitleaks v3); added `codeql.yml`, `security.yml` (golangci-lint + govulncheck +
  dependency-review), `dependabot.yml`, `.golangci.yml` (depguard stdlib-only + gosec).
- **Release:** GoReleaser (`.goreleaser.yaml` + tag-triggered `release.yml`) with SLSA
  build-provenance attestation; decoupled release from `build-hashctl.yml`.
- **Architecture-driven design:** 7 MADR ADRs + index, `docs/threat-model.md` (STRIDE),
  `doc.go` package docs.
- **Bumped `go.mod` to `go 1.25.0`** (supported), replacing end-of-life `go 1.19`; validated
  with go1.25.11 (the toolchain CI's setup-go installs).
- **Copilot:** preserved the Memory Bank framework, added project instructions, this
  populated `/memory-bank/`, four agent skills, two prompt files, and
  `copilot-setup-steps.yml`.

## Next steps / active decisions

- Enforce the `main` ruleset: required PR review + CODEOWNERS + signed commits + linear
  history (a GitHub settings change, not a file). Retire/refresh `.github/COMPENSATING_CONTROLS.md`.
- Cut the first release tag `v0.1.0` once the branch merges.
- Optional follow-ups: SBOM on release, `go test -fuzz` schedule, a Diátaxis `docs/` tree.
