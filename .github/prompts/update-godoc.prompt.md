---
mode: agent
description: Update godoc comments and README/docs after a change to hashctl.
---

# Update hashctl documentation

Update documentation for recent changes to `${input:target:Which file, command, or package changed?}`.

## Standards

1. **Package docs (`doc.go`).** Keep `internal/hashctl/doc.go` and `cmd/hashctl/doc.go`
   accurate. The command package comment must begin with "Hashctl is a command-line
   client …"; the library package comment must begin with "Package hashctl …". Reference
   exported symbols with doc links (e.g. `[Run]`, `[ExitUsage]`).
2. **Exported symbols.** A godoc comment starts with the symbol name.
3. **README.md.** If a subcommand, flag, env var, or exit code changed, update the Usage,
   Configuration, and Exit-codes sections so they match `internal/hashctl/help.go` and
   `types.go` exactly.
4. **Config precedence.** If `resolveConfig` changed, update the precedence description
   (defaults → config file → env → CLI flags) in the README and `doc.go` (ADR-0005).
5. **CHANGELOG.md.** Add an `[Unreleased]` entry for any user-facing change.
6. **ADRs.** If the change alters an architectural decision, add or supersede an ADR in
   `docs/decisions/`.

## Constraints

- Do not modify test files as part of a documentation update.
- Do not add third-party documentation generators or any dependency.
- Keep docs consistent with the code — verify command/flag names against `help.go`.
