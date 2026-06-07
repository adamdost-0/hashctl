# Product Context

> Why hashctl exists, the problems it solves, and its user-experience goals.

## Why it exists

The Hash Engine API drives a cryptographic CTS manifest signing workflow: a job hashes
source blobs, a manifest is built, and two signatures are applied. Operators and CI
pipelines need a safe, scriptable way to drive that workflow without hand-crafting HTTP
requests or risking credential leakage.

## Problems it solves

- **Safe credential handling.** Bearer tokens and Azure storage credentials are easy to
  leak in shell history, logs, or error output. hashctl refuses literal token flags,
  enforces `chmod 600` token files, and redacts secrets from all output.
- **Transport safety.** It refuses to send bearer tokens over non-loopback plaintext HTTP
  and requires HTTPS for remote hosts.
- **Automation ergonomics.** Semantic exit codes and a JSON output mode let CI branch on
  outcomes deterministically; polling helpers wait for jobs to reach a target state.

## User experience goals

- **Two audiences:** interactive operators (default `human` output, concise summaries) and
  automation (`--output json`, stable shapes, semantic exit codes).
- **Predictable configuration:** defaults → config file → environment → CLI flags.
- **Discoverable:** `hashctl help` and `hashctl <command> --help` document every command.
- **Trustworthy releases:** checksummed, attested artifacts installable via a helper script.
