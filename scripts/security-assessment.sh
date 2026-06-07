#!/usr/bin/env bash
# security-assessment.sh — collect a machine-readable static-analysis evidence
# bundle for the hashctl multi-model security assessment.
#
# Stdlib/standard-tooling only: this script shells out to the Go toolchain and
# optional security scanners if they are installed. It NEVER adds Go dependencies,
# never edits go.mod, and runs every Go command offline (GOPROXY=off GOSUMDB=off).
# It degrades gracefully: a missing tool is recorded as "skip", not a hard error.
#
# Output: a single JSON object on stdout (or to --output FILE) describing each
# check as pass | fail | skip. Exit status is 0 unless --strict is given and at
# least one check failed.
#
# Usage:
#   scripts/security-assessment.sh [--output FILE] [--strict]
set -uo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${repo_root}"

out_file=""
strict=0
while [[ $# -gt 0 ]]; do
  case "$1" in
    --output) out_file="${2:-}"; shift 2 ;;
    --strict) strict=1; shift ;;
    -h|--help)
      grep '^#' "${BASH_SOURCE[0]}" | sed 's/^# \{0,1\}//'
      exit 0 ;;
    *) echo "unknown argument: $1" >&2; exit 2 ;;
  esac
done

export CGO_ENABLED=0 GOPROXY=off GOSUMDB=off
# Use the locally installed toolchain only; never attempt an offline download.
export GOTOOLCHAIN="${GOTOOLCHAIN:-local}"

pass=0
fail=0
skip=0
checks_json=""

json_escape() { # stdin -> JSON-safe single-line string (no surrounding quotes), truncated
  local s
  s="$(cat)"
  s="${s//\\/\\\\}"
  s="${s//\"/\\\"}"
  s="${s//$'\t'/ }"
  s="${s//$'\r'/}"
  s="${s//$'\n'/ }"
  if [[ ${#s} -gt 400 ]]; then s="${s:0:400}..."; fi
  printf '%s' "$s"
}

add_check() { # name available(true|false) status(pass|fail|skip) detail
  local name="$1" available="$2" status="$3" detail="$4"
  local sep=""
  [[ -n "${checks_json}" ]] && sep=","
  checks_json="${checks_json}${sep}{\"name\":\"${name}\",\"available\":${available},\"status\":\"${status}\",\"detail\":\"${detail}\"}"
  case "${status}" in
    pass) pass=$((pass + 1)) ;;
    fail) fail=$((fail + 1)) ;;
    skip) skip=$((skip + 1)) ;;
  esac
}

have() { command -v "$1" >/dev/null 2>&1; }

RC=0
OUT=""
run_capture() { OUT="$("$@" 2>&1)"; RC=$?; return 0; }

# --- Go toolchain probe ----------------------------------------------------
go_available=false
go_version=""
go_toolchain_ok=false
required_go="$(awk '/^go [0-9]/{print $2; exit}' go.mod 2>/dev/null || true)"
if have go; then
  go_available=true
  go_version="$(go version 2>/dev/null | awk '{print $3}')"
fi

toolchain_too_old() { # detect "go.mod requires go >= X" style messages
  [[ "$1" == *"requires go >="* ]] || [[ "$1" == *"go.mod requires"* ]]
}

# --- gofmt -----------------------------------------------------------------
if have gofmt; then
  run_capture gofmt -l .
  if [[ ${RC} -eq 0 && -z "${OUT}" ]]; then
    add_check "gofmt" true "pass" "no formatting differences"
  elif [[ ${RC} -eq 0 ]]; then
    add_check "gofmt" true "fail" "$(printf '%s' "unformatted: ${OUT}" | json_escape)"
  else
    add_check "gofmt" true "fail" "$(printf '%s' "${OUT}" | json_escape)"
  fi
else
  add_check "gofmt" false "skip" "gofmt not installed"
fi

# --- go vet ----------------------------------------------------------------
if [[ "${go_available}" == true ]]; then
  run_capture go vet ./...
  if toolchain_too_old "${OUT}"; then
    add_check "go vet" true "skip" "$(printf '%s' "local toolchain ${go_version} < required go ${required_go}; covered in CI" | json_escape)"
  elif [[ ${RC} -eq 0 ]]; then
    go_toolchain_ok=true
    add_check "go vet" true "pass" "no vet diagnostics"
  else
    add_check "go vet" true "fail" "$(printf '%s' "${OUT}" | json_escape)"
  fi
else
  add_check "go vet" false "skip" "go not installed"
fi

# --- go test ---------------------------------------------------------------
if [[ "${go_available}" == true ]]; then
  run_capture go test ./...
  if toolchain_too_old "${OUT}"; then
    add_check "go test" true "skip" "$(printf '%s' "local toolchain ${go_version} < required go ${required_go}; covered in CI" | json_escape)"
  elif [[ ${RC} -eq 0 ]]; then
    add_check "go test" true "pass" "all packages passed"
  else
    add_check "go test" true "fail" "$(printf '%s' "${OUT}" | json_escape)"
  fi
else
  add_check "go test" false "skip" "go not installed"
fi

# --- go.mod dependency policy (ADR-0001: stdlib-only, no go.sum) ------------
require_count="$(grep -cE '^[[:space:]]*require' go.mod 2>/dev/null || true)"
require_count="${require_count:-0}"
if [[ "${require_count}" -eq 0 && ! -f go.sum ]]; then
  add_check "stdlib-only" true "pass" "0 require entries, no go.sum"
else
  add_check "stdlib-only" true "fail" "$(printf '%s' "require entries=${require_count}; go.sum present=$([[ -f go.sum ]] && echo yes || echo no)" | json_escape)"
fi

# --- govulncheck -----------------------------------------------------------
if have govulncheck; then
  run_capture govulncheck ./...
  if toolchain_too_old "${OUT}"; then
    add_check "govulncheck" true "skip" "$(printf '%s' "toolchain too old; covered in CI" | json_escape)"
  elif [[ ${RC} -eq 0 ]]; then
    add_check "govulncheck" true "pass" "no known vulnerabilities"
  else
    add_check "govulncheck" true "fail" "$(printf '%s' "${OUT}" | json_escape)"
  fi
else
  add_check "govulncheck" false "skip" "govulncheck not installed (runs in security.yml)"
fi

# --- gosec / golangci-lint -------------------------------------------------
if have golangci-lint; then
  run_capture golangci-lint run ./...
  if [[ ${RC} -eq 0 ]]; then
    add_check "golangci-lint" true "pass" "0 issues"
  else
    add_check "golangci-lint" true "fail" "$(printf '%s' "${OUT}" | json_escape)"
  fi
elif have gosec; then
  run_capture gosec ./...
  if [[ ${RC} -eq 0 ]]; then
    add_check "gosec" true "pass" "0 issues"
  else
    add_check "gosec" true "fail" "$(printf '%s' "${OUT}" | json_escape)"
  fi
else
  add_check "golangci-lint" false "skip" "golangci-lint/gosec not installed (runs in security.yml)"
fi

# --- gitleaks --------------------------------------------------------------
if have gitleaks; then
  run_capture gitleaks detect --no-banner --redact
  if [[ ${RC} -eq 0 ]]; then
    add_check "gitleaks" true "pass" "no leaks detected"
  else
    add_check "gitleaks" true "fail" "$(printf '%s' "potential secret(s) detected; review (output redacted)" | json_escape)"
  fi
else
  add_check "gitleaks" false "skip" "gitleaks not installed (runs in build-hashctl.yml)"
fi

# --- emit JSON -------------------------------------------------------------
generated_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
document="$(cat <<JSON
{
  "tool": "security-assessment.sh",
  "schema_version": 1,
  "generated_at": "${generated_at}",
  "repo": "adamdost-0/hashctl",
  "go": {"available": ${go_available}, "version": "${go_version}", "required": "${required_go}", "toolchain_ok": ${go_toolchain_ok}},
  "checks": [${checks_json}],
  "summary": {"pass": ${pass}, "fail": ${fail}, "skip": ${skip}}
}
JSON
)"

if [[ -n "${out_file}" ]]; then
  printf '%s\n' "${document}" >"${out_file}"
  echo "wrote evidence bundle to ${out_file}" >&2
else
  printf '%s\n' "${document}"
fi

if [[ "${strict}" -eq 1 && "${fail}" -gt 0 ]]; then
  exit 1
fi
exit 0
