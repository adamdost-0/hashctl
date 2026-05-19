#!/usr/bin/env bash
set -euo pipefail

CGO_ENABLED=0 GOPROXY=off GOSUMDB=off go test ./...
