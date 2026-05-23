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
if [[ ! -f "${binary}" ]]; then
  echo "Expected binary not found at ${binary}" >&2
  exit 1
fi

file_output="$(file "${binary}")"
case "${platform}" in
  linux-amd64)
    if [[ "${file_output}" != *"ELF 64-bit LSB executable"* ]]; then
      echo "Expected linux-amd64 executable metadata, got: ${file_output}" >&2
      exit 1
    fi
    ;;
  darwin-arm64)
    if [[ "${file_output}" != *"Mach-O 64-bit"* ]] || [[ "${file_output}" != *"arm64"* ]]; then
      echo "Expected darwin-arm64 executable metadata, got: ${file_output}" >&2
      exit 1
    fi
    ;;
  *)
    echo "Unsupported platform '${platform}'" >&2
    exit 1
    ;;
esac

if ! strings "${binary}" | grep -F "${version}" >/dev/null; then
  echo "Expected embedded version '${version}' not found in ${binary}" >&2
  exit 1
fi

echo "Verified ${artifact} from ${release_tag}"
