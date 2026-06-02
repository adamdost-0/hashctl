#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo="${HASHCTL_REPO:-adamdost-0/hashctl}"
version="${HASHCTL_VERSION:-$("${script_dir}/get-hashctl-version.sh")}"
platform="${HASHCTL_PLATFORM:-linux-amd64}"
install_dir="${HASHCTL_INSTALL_DIR:-/usr/local/bin}"
release_tag="${HASHCTL_RELEASE_TAG:-v${version}}"

if [[ -z "${version}" ]]; then
  echo "Unable to determine hashctl version" >&2
  exit 1
fi

artifact="hashctl-${version}-${platform}.tar.gz"
checksum="${artifact}.sha256"

tmp_dir="$(mktemp -d)"
trap 'rm -rf "${tmp_dir}"' EXIT

if command -v gh >/dev/null 2>&1; then
  gh release download "${release_tag}" \
    --repo "${repo}" \
    --pattern "${artifact}" \
    --pattern "${checksum}" \
    --dir "${tmp_dir}"
else
  curl -fsSL "https://github.com/${repo}/releases/download/${release_tag}/${artifact}" \
    -o "${tmp_dir}/${artifact}"
  curl -fsSL "https://github.com/${repo}/releases/download/${release_tag}/${checksum}" \
    -o "${tmp_dir}/${checksum}"
fi

(
  cd "${tmp_dir}"
  shasum -a 256 -c "${checksum}"
  tar -xzf "${artifact}"
)

binary="${tmp_dir}/hashctl-${version}-${platform}/hashctl"
if [[ ! -f "${binary}" ]]; then
  echo "Expected binary not found at ${binary}" >&2
  exit 1
fi

mkdir -p "${install_dir}"
install_cmd=(install)
if [[ ! -w "${install_dir}" ]]; then
  if command -v sudo >/dev/null 2>&1; then
    install_cmd=(sudo install)
  else
    echo "No write access to ${install_dir}; rerun with sudo or set HASHCTL_INSTALL_DIR to a writable directory" >&2
    exit 1
  fi
fi

"${install_cmd[@]}" -m 0755 "${binary}" "${install_dir}/hashctl"

printf 'Installed hashctl to %s/hashctl\n' "${install_dir}"
"${install_dir}/hashctl" version
