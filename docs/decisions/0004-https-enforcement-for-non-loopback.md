---
status: accepted
date: 2026-06-07
decision-makers: [adamdost-0]
---

# HTTPS Enforcement for Non-Loopback Hosts

## Context and Problem Statement

`hashctl` sends bearer tokens and signing requests to the Hash Engine API. Plaintext HTTP
to a remote host would expose credentials to network interception. How strictly should the
client enforce transport security, while still allowing local development and smoke tests?

## Decision Drivers

- Bearer tokens must never traverse the network unencrypted to a remote host.
- Local development and smoke tests against `localhost` need to work without TLS.
- The control should be defense-in-depth, not a single check.

## Considered Options

- **Require HTTPS for all non-loopback hosts; allow HTTP only for loopback.**
- **Require HTTPS unconditionally** (blocks local development over HTTP).
- **Warn but allow** plaintext HTTP.

## Decision Outcome

Chosen option: **require HTTPS for non-loopback hosts**. The API URL is validated at config
resolution time (`config.go`), and the bearer token is independently refused over
non-loopback plaintext HTTP at request time (`client.go applyHeaders`). Loopback hosts
(`localhost`, `127.0.0.1`, `::1`) may use HTTP.

### Consequences

- Good: credentials cannot be sent in cleartext to a remote endpoint.
- Good: two independent checks (config-time and request-time) provide defense in depth.
- Good: local development and smoke flows over `http://localhost` still work.
- Bad: operators behind a TLS-terminating proxy on a non-loopback address must front it
  with HTTPS.

### Confirmation

This is a knock-out criterion: no option may relax HTTPS for non-loopback hosts.
`commands_test.go` covers both the rejection of non-loopback HTTP and the loopback-HTTP
allowance.
