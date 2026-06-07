#!/usr/bin/env bash
# sandbox.sh — build and test hashctl inside a Go 1.25 container, host-agnostic.
#
# The host needs only Docker; no Go toolchain is required. The repository is
# bind-mounted, commands run as the host user (no root-owned artifacts), and the
# Go environment matches CI and hashctl's offline, stdlib-only policy
# (CGO_ENABLED=0, GOPROXY=off, GOSUMDB=off).
#
# Usage:
#   scripts/sandbox.sh                # run the validation gate (vet, test, gofmt, build)
#   scripts/sandbox.sh go test ./...  # run an arbitrary command in the sandbox
#   scripts/sandbox.sh gate           # explicit alias for the validation gate
#
# Environment:
#   HASHCTL_SANDBOX_IMAGE   container image to use (default: golang:1.25)
set -euo pipefail

image="${HASHCTL_SANDBOX_IMAGE:-golang:1.25}"
repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if ! command -v docker >/dev/null 2>&1; then
  echo "sandbox.sh: docker is required but not found on PATH" >&2
  exit 127
fi
if ! docker info >/dev/null 2>&1; then
  echo "sandbox.sh: the docker daemon is not responding (is it running / do you have permission?)" >&2
  exit 1
fi

gate_cmd='set -e
echo "# $(go version)"
echo "## go vet"
go vet ./...
echo "## go test"
go test ./...
echo "## gofmt"
unformatted="$(gofmt -l .)"
if [ -n "$unformatted" ]; then echo "unformatted files:"; echo "$unformatted"; exit 1; fi
echo "gofmt clean"
echo "## go build"
go build -o /tmp/hashctl ./cmd/hashctl
echo "build OK"'

if [[ $# -eq 0 || "${1:-}" == "gate" ]]; then
  run_cmd="${gate_cmd}"
else
  run_cmd="$*"
fi

exec docker run --rm \
  -v "${repo_root}":/work -w /work \
  --user "$(id -u):$(id -g)" \
  -e HOME=/tmp -e GOCACHE=/tmp/gocache -e GOPATH=/tmp/gopath \
  -e CGO_ENABLED=0 -e GOPROXY=off -e GOSUMDB=off \
  "${image}" bash -c "${run_cmd}"
