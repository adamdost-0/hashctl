<!-- Describe your change and link the issue it addresses, e.g. "Fixes #123". -->

## Summary

## Checklist

- [ ] Commit messages follow [Conventional Commits](https://www.conventionalcommits.org/)
- [ ] Validation gate passes locally: `test → vet → gofmt → build`
- [ ] `golangci-lint run` passes (if installed)
- [ ] No third-party dependencies added (`go.mod` has no `require` entries)
- [ ] New user-facing output is routed through `redact()`
- [ ] `CHANGELOG.md` updated under `[Unreleased]` (for user-facing changes)
- [ ] Docs updated where relevant (`README.md`, `docs/`, `doc.go`)
