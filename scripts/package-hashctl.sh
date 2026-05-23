#!/usr/bin/env bash
set -euo pipefail

version="${HASHCTL_VERSION:-dev}"
os_name="${GOOS:-$(go env GOOS)}"
arch_name="${GOARCH:-$(go env GOARCH)}"
artifact_dir="artifacts/hashctl"
package_name="hashctl-${version}-${os_name}-${arch_name}"
archive_name="${package_name}.tar.gz"
ldflags="-X github.com/adamdost-0/hash-engine/internal/hashctl.Version=${version}"

mkdir -p "${artifact_dir}/${package_name}"
GOPROXY=off GOSUMDB=off go build -ldflags "${ldflags}" -o "${artifact_dir}/${package_name}/hashctl" ./cmd/hashctl
chmod 0755 "${artifact_dir}/${package_name}/hashctl"
tar -C "${artifact_dir}" -czf "${artifact_dir}/${archive_name}" "${package_name}"
(
  cd "${artifact_dir}"
  shasum -a 256 "${archive_name}" > "${archive_name}.sha256"
)
