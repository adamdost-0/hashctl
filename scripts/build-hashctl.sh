#!/usr/bin/env bash
set -euo pipefail

mkdir -p bin
GOPROXY=off GOSUMDB=off go build -o bin/hashctl ./cmd/hashctl
