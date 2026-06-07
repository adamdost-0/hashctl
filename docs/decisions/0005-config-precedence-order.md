---
status: accepted
date: 2026-06-07
decision-makers: [adamdost-0]
---

# Configuration Precedence Order

## Context and Problem Statement

`hashctl` settings can come from built-in defaults, a config file, environment variables,
and CLI flags. When more than one source supplies the same setting, which one wins?

## Decision Drivers

- Predictability: the same inputs must always resolve to the same configuration.
- Operability: operators expect explicit flags to override ambient configuration.
- Security: secrets should come from the environment or a protected file, not be baked
  into a committed config file.

## Considered Options

- **defaults → config file → environment → CLI flags** (later wins).
- CLI flags lowest precedence (flags overridden by env) — counterintuitive.
- Single source only (e.g. flags-only) — inflexible for CI and local use.

## Decision Outcome

Chosen option: **defaults → config file → environment → CLI flags**, resolved by
`resolveConfig` in `internal/hashctl/config.go`. Only `api_url` and `output` are read from
the config file; secrets are never persisted to it.

### Consequences

- Good: explicit `--flags` always win, which matches operator expectations.
- Good: secrets stay in `HASH_ENGINE_BEARER_TOKEN` or a `chmod 600` token file.
- Bad: four layers must be kept consistent between code and documentation.

### Confirmation

`commands_test.go` integration tests drive precedence with fake environment maps. The
README and `doc.go` precedence description must match the order implemented in
`resolveConfig`.
