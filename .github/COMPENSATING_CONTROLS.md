# Compensating Controls (C-3)

- Branch protection/rulesets are **not active yet** for this repository.
- C-3 compensating control for this interim state is enforced via required CI checks in pull requests and push validation (`build-hashctl.yml`), including:
  - Go test/vet/lint/build validation
  - Credential scanning with gitleaks
- Final branch-protection cutover remains blocked until repository rulesets are configured and enforced.

- **Release Integrity:** Release tags must be annotated and signed. This is enforced by `.github/workflows/release.yml` and should be additionally protected by a tag ruleset once branch protection cutover occurs.
