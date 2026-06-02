#!/usr/bin/env bash
set -euo pipefail

rm -rf bin
mkdir -p bin
version="${HASHCTL_VERSION:-dev}"
ldflags="-X github.com/adamdost-0/hashctl/internal/hashctl.Version=${version}"
GOPROXY=off GOSUMDB=off go build -ldflags "${ldflags}" -o bin/hashctl ./cmd/hashctl
