# Project Brief: hashctl

> Foundation document for the Memory Bank. Defines core requirements, goals, and scope.

## What it is

`hashctl` is a command-line client for the **Hash Engine** REST API. It manages CTS
manifest hashing **jobs** and the two-phase **signing** workflow, and reports liveness via
`health`. It is written in Go using **only the standard library**.

## Goals

- A small, auditable, dependency-free CLI suitable for operators and CI automation.
- Deterministic, scriptable behavior: semantic exit codes and a machine-readable JSON mode.
- Security-by-construction: credentials never leak to output; transport is HTTPS for
  non-loopback hosts; tokens come only from the environment or a `chmod 600` file.
- Reproducible, offline builds and signed, attested releases.

## Scope

In scope: `health`, `job` (create/get/list/cancel), `manifest` (build/get),
`sign` (first/second), `smoke` (single-job/multi-job), plus `help` and `version`;
human and JSON output; configuration via flags, environment, and a `config.json`.

Out of scope: a public Go API (everything is `package main` + `internal/`); third-party
dependencies; output formats other than human and JSON; key management (delegated to the
Hash Engine / Key Vault).

## Success criteria

The validation gate (test → vet → gofmt → build) and the CI security gates pass; releases
are reproducible and attested; no credential ever appears unredacted in output or logs.
