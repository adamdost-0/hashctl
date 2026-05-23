#!/usr/bin/env bash
set -euo pipefail

version_file="${1:-VERSION}"

if [[ ! -f "${version_file}" ]]; then
  echo "Version file not found: ${version_file}" >&2
  exit 1
fi

tr -d '\r\n' < "${version_file}"
