# Branch Protection & Governance Controls

The `main` branch is protected by a repository ruleset enforcing the following controls:

## Enforced ruleset

| Control | Enforcement |
|---|---|
| **Required PR review** | At least one approval from a CODEOWNERS entry before merge |
| **Required status checks** | `build-hashctl.yml` (test/vet/gofmt/build), `codeql.yml` (SAST), `security.yml` (golangci-lint, govulncheck, dependency-review) |
| **Signed commits** | All commits on `main` must carry a verified GPG or SSH signature |
| **Linear history** | Squash or rebase merges only; direct force-push to `main` is forbidden |
| **Tag protection** | Version tags (`v*`) may only be created by repository administrators |

## Credential scanning

Gitleaks runs on every PR and push via `build-hashctl.yml`. Secrets found in history are
treated as an incident: remove from history and rotate the affected credential immediately.

## Reference

See [`CONTRIBUTING.md`](../CONTRIBUTING.md) for the PR workflow that these controls enforce.
See [`SECURITY.md`](../SECURITY.md) for the vulnerability disclosure process.

