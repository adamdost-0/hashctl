# Architecture Decision Records

This directory captures the significant architectural decisions for `hashctl` using the
[MADR](https://adr.github.io/madr/) (Markdown Any Decision Records) format.

Each record is immutable once accepted; to change a decision, add a new ADR that
supersedes the old one (and update the old record's status).

## Index

| ADR | Title | Status |
|---|---|---|
| [0001](0001-stdlib-only-zero-dependency.md) | Stdlib-only, zero external dependency policy | Accepted |
| [0002](0002-semantic-exit-codes.md) | Semantic exit codes | Accepted |
| [0003](0003-credential-redaction-on-all-output.md) | Credential redaction on all output paths | Accepted |
| [0004](0004-https-enforcement-for-non-loopback.md) | HTTPS enforcement for non-loopback hosts | Accepted |
| [0005](0005-config-precedence-order.md) | Configuration precedence order | Accepted |
| [0006](0006-internal-package-boundary.md) | All logic in the `internal/hashctl` package | Accepted |
| [0007](0007-dual-output-modes.md) | Two output modes only (human and JSON) | Accepted |

## Writing a new ADR

1. Copy the structure of an existing record.
2. Name it `NNNN-short-title.md` (zero-padded, next number).
3. Set `status: proposed`, fill in the sections, and open a PR.
4. On merge, set `status: accepted` and add it to the index above.
