---
status: accepted
date: 2026-06-07
decision-makers: [adamdost-0]
---

# Two Output Modes Only (Human and JSON)

## Context and Problem Statement

`hashctl` serves both interactive operators (who want readable summaries) and automation
(which wants machine-parseable output). How many and which output formats should the CLI
support?

## Decision Drivers

- Operators need concise, readable output by default.
- Automation needs stable, parseable output.
- Each additional format multiplies rendering and test surface.

## Considered Options

- **Two modes: `human` (default) and `json`.**
- Add more formats (YAML, table, template) on demand.
- JSON only (poor interactive experience).

## Decision Outcome

Chosen option: **exactly two modes — `human` and `json`** — validated by `resolveConfig`,
which rejects any other value. `writeHuman` type-switches on the result type; `writeJSON`
emits structured output. Both paths route through `redact()`.

### Consequences

- Good: small, well-tested rendering surface.
- Good: invalid `--output` values fail fast with a usage error.
- Bad: adding a result type requires a new `case` in `writeHuman`.
- Bad: a third format would require a new ADR superseding this one.

### Confirmation

`resolveConfig` validation acts as the fitness function (only `human`/`json` accepted).
Command tests assert both the human summary and the JSON shape per command.
