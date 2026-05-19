#!/usr/bin/env bash
set -euo pipefail

GOPROXY=off GOSUMDB=off go test ./...
