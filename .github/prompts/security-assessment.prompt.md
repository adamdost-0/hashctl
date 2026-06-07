---
mode: agent
description: Run a multi-model security assessment of hashctl with parallel sub-agents and synthesize findings into GitHub issues.
---

# Multi-model security assessment

Assess `hashctl` across three dimensions — **repo structure**, **repo health**, and
**code security** (the CLI command surface) — using several models in parallel, then file
the results as GitHub issues. hashctl is stdlib-only: this workflow adds **no** Go
dependencies and never edits `go.mod`.

## 1. Baseline

Run `scripts/security-assessment.sh --output <bundle.json>` to collect a machine-readable
static-analysis evidence bundle (gofmt / `go vet` / `go test`, the stdlib-only policy, and
`govulncheck` / `gosec` / `gitleaks` where installed; missing tools degrade to `skip`, not
an error). Share the bundle with every sub-agent. Create triage labels: `security`,
`severity:{critical,high,medium,low,info}`, `area:{structure,health,code}`.

## 2. Parallel multi-model panel

Dispatch parallel `task` sub-agents (`general-purpose`), each with a different `model`
override so **≥3 vendors** participate. Partition by command/area and **double-cover the
credential and redaction crown jewels** with two vendors. Suggested panel:

| Unit | Scope | Vendor |
|---|---|---|
| structure | layout, `internal/` boundary, ADR enforcement | Google (gemini) |
| health | workflows, supply chain, governance | OpenAI (gpt) |
| credentials + transport | `config.go`, `client.go` `applyHeaders`/`do` | Anthropic (opus) |
| redaction + output | `output.go` + every success-render path | OpenAI (gpt-codex) |
| job + manifest | `commands.go` job/manifest, `client.go` methods | Google (gemini) |
| sign + smoke + health | `commands.go` sign/health, `smoke.go` | Anthropic (sonnet) |
| argparse + exit codes | `parseGlobal`, dispatch, `errorExitCode` | OpenAI (gpt) |
| dataflow (adversarial) | untrusted response → output, credential lifecycle | Google (gemini) |

Each sub-agent is **READ-ONLY**, must not emit secrets (the repo is public), and returns
findings in the schema below. Point each at the baseline bundle and at `docs/decisions/`
(ADRs) and `docs/threat-model.md` (STRIDE).

## 3. Finding schema (strict JSON, one object per finding)

`finding_id · area · title · severity {Critical|High|Medium|Low|Info} · stride · evidence
(file:line) · analysis · recommendation · confidence · model · adr_ref`

Severity: **Critical** = secret leak / RCE / auth bypass shippable to a customer;
**High** = likely exploitable invariant break; **Medium** = defense-in-depth or validation
gap; **Low** = hardening; **Info** = observation.

## 4. Primary synthesis

The primary model dedupes and merges findings across models (convergent findings raise
confidence), validates each against source (cite `file:line`), normalizes severity, maps
to STRIDE + the ADRs, and writes one prioritized, human-readable report.

## 5. GitHub issues

Secret-scrub every body. **Critical/High/Medium** → individual issues (labels `security`,
`severity:*`, `area:*`); **Low/Info** → one rolled-up summary issue; plus one tracking
issue linking them all. Dedupe against existing issues before filing. Keep bodies
resolution-focused — no real secrets or weaponized exploit detail.
