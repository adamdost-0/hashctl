# Compensating Controls (C-3)

- Branch protection/rulesets are **not active yet** for this repository.
- C-3 compensating control for this interim state is enforced via required CI checks in pull requests and push validation (`build-hashctl.yml`), including:
  - Go test/vet/lint/build validation
  - Credential scanning with gitleaks
- Final branch-protection cutover remains blocked until repository rulesets are configured and enforced.
