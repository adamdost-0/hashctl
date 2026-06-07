---
name: release-notes-writer
description: Prepare a hashctl release. Use when asked to draft release notes, update the changelog, bump the version, or cut a release. Determines the semver bump from commits and updates VERSION and CHANGELOG.
---

# hashctl Release Notes Writer

hashctl follows Semantic Versioning and Keep a Changelog. Releases are cut from an
annotated, signed tag `vX.Y.Z`; `release.yml` (GoReleaser) builds and publishes from the
tag. The `VERSION` file must match the tag (release.yml enforces this).

## Step 1 — Determine the semver bump

Read commits since the last tag: `git log $(git describe --tags --abbrev=0 2>/dev/null)..HEAD --oneline`
(if there is no previous tag, review all commits). Apply Conventional Commits:

- **MAJOR:** any `feat!:` / `BREAKING CHANGE:` — removed/renamed commands or flags,
  changed exit codes, changed output contract.
- **MINOR:** `feat:` — new subcommands, flags, output fields, or API support.
- **PATCH:** `fix:` and security fixes; `chore`/`ci`/`docs`/`test`/`refactor` alone do not
  bump.

While `0.y.z`, breaking changes bump MINOR, not MAJOR (the API is not yet stable).

## Step 2 — Update CHANGELOG.md

Move entries from `[Unreleased]` into a new `## [X.Y.Z] - YYYY-MM-DD` section using the
Keep a Changelog categories (Added, Changed, Deprecated, Removed, Fixed, Security). Update
the `[Unreleased]` and `[X.Y.Z]` compare/tag links at the bottom.

## Step 3 — Update VERSION

Write the new version (no `v` prefix) to the `VERSION` file.

## Step 4 — Commit and tag

```bash
git commit -am "chore(release): v<X.Y.Z>"
git tag -s "v<X.Y.Z>" -m "hashctl v<X.Y.Z>"   # signed, annotated
git push origin main "v<X.Y.Z>"               # release.yml builds and publishes
```

## Constraints

- Use `git` and shell only — no third-party changelog tools, no dependencies.
- Do not hand-upload assets; GoReleaser (release.yml) owns the GitHub Release.
