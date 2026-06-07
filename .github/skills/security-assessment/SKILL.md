---
name: security-assessment
description: Run a multi-model (cross-vendor) security assessment of hashctl. Use when asked to security-assess, run a multi-model/multi-modal security review, evaluate repo structure/health/code-security in parallel, or spawn GitHub issues from a security audit. Delegates per-command analysis to parallel sub-agents (one model each) and synthesizes the findings.
---

# hashctl Multi-Model Security Assessment

Orchestrates a cross-vendor security assessment of `hashctl` and files the results as
GitHub issues. It complements the single-pass `security-review` skill by running several
models in parallel and synthesizing their findings. hashctl is stdlib-only — this workflow
adds **no** Go dependencies and never edits `go.mod`; its tooling is shell + markdown.

## When to use

- "Run a multi-model / multi-modal security assessment", "assess repo structure, health,
  and code security in parallel", "have several models evaluate the commands and open
  issues with analysis and resolution recommendations".

## Workflow

1. **Baseline.** `scripts/security-assessment.sh --output <bundle.json>` → a machine-readable
   evidence bundle (gofmt / `go vet` / `go test`, the stdlib-only policy, and
   `govulncheck` / `gosec` / `gitleaks` if installed; missing tools become `skip`, never a
   hard error). Create triage labels (`security`, `severity:{critical,high,medium,low,info}`,
   `area:{structure,health,code}`).
2. **Parallel panel.** Dispatch parallel `task` sub-agents (`general-purpose`), each with a
   different `model` so ≥3 vendors participate; partition by command/area and double-cover
   the credential + redaction paths. See `.github/prompts/security-assessment.prompt.md`
   for the panel matrix and the strict finding schema. Sub-agents are READ-ONLY and must
   not emit secrets (the repo is public).
3. **Synthesize.** The primary model dedupes/merges findings across models (convergence
   raises confidence), validates each against source (`file:line`), normalizes severity,
   maps to STRIDE and the ADRs in `docs/decisions/` + `docs/threat-model.md`, and writes
   one prioritized report.
4. **File issues.** Secret-scrub every body. Critical/High/Medium → individual issues;
   Low/Info → one summary issue; plus a tracking issue linking them. Dedupe against
   existing issues.

## Severity

Critical = credential leak / RCE / auth bypass shippable to a customer; High = likely
exploitable invariant violation; Medium = defense-in-depth or validation gap; Low =
hardening; Info = observation.

## Guardrails

- READ-ONLY on code: the assessment creates issues; it does not change source.
- Public repo: never put real secrets, tokens, or weaponized exploit detail in issue bodies.
- Preserve the stdlib-only, offline invariants (ADR-0001); keep the credential, redaction,
  and HTTPS controls (ADR-0003/0004) front-of-mind when judging severity.
- Record durable invariants discovered during the assessment with `store_memory`, and
  update the Memory Bank (`activeContext.md`, `progress.md`).
