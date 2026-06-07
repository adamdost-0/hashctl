# Compensating Controls (C-3)

- Branch protection/rulesets are **not active yet** for this repository.
- C-3 compensating control for this interim state is enforced via required CI checks in pull requests and push validation (`build-hashctl.yml`), including:
  - Go test/vet/lint/build validation
  - Credential scanning with gitleaks
- Final branch-protection cutover remains blocked until repository rulesets are configured and enforced.

- **Release Integrity:** Release tags must be annotated and signed. This is enforced by `.github/workflows/release.yml` and should be additionally protected by a tag ruleset once branch protection cutover occurs.

## Release-signing key setup (prerequisite before the first release)

The `release.yml` workflow verifies the pushed tag with `git verify-tag` and fails
closed if the tag is not annotated **and** signed. Before cutting any release,
configure **one** of the following repository secrets so the runner can verify the
signer's public key (without it, every release push fails at the verification step):

- **SSH-signed tags (recommended):** set `ALLOWED_SIGNERS` to a git allowed-signers
  line for the release signer, e.g.
  `release@example.com namespaces="git" ssh-ed25519 AAAA...`
- **GPG-signed tags:** set `TAG_SIGNING_PUBKEY` to the ASCII-armored public key of the
  signer (`gpg --armor --export <key-id>`).

Create release tags with `git tag -s -a vX.Y.Z -m "vX.Y.Z"` (annotated **and** signed).
