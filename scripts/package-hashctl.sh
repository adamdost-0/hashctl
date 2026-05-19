#!/usr/bin/env bash
set -euo pipefail

version="${HASHCTL_VERSION:-dev}"
os_name="${GOOS:-$(go env GOOS)}"
arch_name="${GOARCH:-$(go env GOARCH)}"
artifact_dir="artifacts/hashctl"
package_name="hashctl-${version}-${os_name}-${arch_name}"

mkdir -p "${artifact_dir}/${package_name}"
GOPROXY=off GOSUMDB=off go build -o "${artifact_dir}/${package_name}/hashctl" ./cmd/hashctl
tar -C "${artifact_dir}" -czf "${artifact_dir}/${package_name}.tar.gz" "${package_name}"
shasum -a 256 "${artifact_dir}/${package_name}.tar.gz" > "${artifact_dir}/${package_name}.tar.gz.sha256"
