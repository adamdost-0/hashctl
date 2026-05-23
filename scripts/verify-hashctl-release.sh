#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
version="${HASHCTL_VERSION:-$(${script_dir}/get-hashctl-version.sh)}"
platform="${HASHCTL_PLATFORM:-}"
release_tag="${HASHCTL_RELEASE_TAG:-v${version}}"
repo="${HASHCTL_REPO:-adamdost-0/hash-engine}"
download_dir="${HASHCTL_DOWNLOAD_DIR:-}"

if [[ -z "${platform}" ]]; then
  echo "HASHCTL_PLATFORM must be set to a release asset suffix such as linux-amd64" >&2
  exit 1
fi

if [[ -z "${download_dir}" ]]; then
  download_dir="$(mktemp -d)"
  cleanup=1
else
  mkdir -p "${download_dir}"
  cleanup=0
fi

cleanup_dir() {
  if [[ "${cleanup}" -eq 1 ]]; then
    rm -rf "${download_dir}"
  fi
}
trap cleanup_dir EXIT

artifact="hashctl-${version}-${platform}.tar.gz"
checksum="${artifact}.sha256"

gh release download "${release_tag}" \
  --repo "${repo}" \
  --pattern "${artifact}" \
  --pattern "${checksum}" \
  --dir "${download_dir}"

(
  cd "${download_dir}"
  shasum -a 256 -c "${checksum}"
  tar -xzf "${artifact}"
)

binary="${download_dir}/hashctl-${version}-${platform}/hashctl"
if [[ ! -x "${binary}" ]]; then
  echo "Expected binary not found at ${binary}" >&2
  exit 1
fi

version_output="$(${binary} version)"
version_output="${version_output#hashctl }"
if [[ "${version_output}" != "${version}" ]]; then
  echo "Expected version '${version}', got '${version_output}'" >&2
  exit 1
fi

echo "Verified ${artifact} from ${release_tag}"
