# hashctl

> A standard-library-only Go CLI client for the Hash Engine REST API (CTS manifest jobs and signing).

[![Build hashctl](https://github.com/adamdost-0/hashctl/actions/workflows/build-hashctl.yml/badge.svg)](https://github.com/adamdost-0/hashctl/actions/workflows/build-hashctl.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

`hashctl` manages CTS manifest hashing jobs and signing workflows against the Hash Engine
HTTP API. It has **zero third-party dependencies**, builds fully offline, supports both
human-readable and JSON output, and redacts secrets from all output.

## Table of Contents

- [Install](#install)
- [Usage](#usage)
- [Configuration](#configuration)
- [Output modes](#output-modes)
- [Exit codes](#exit-codes)
- [Security](#security)
- [Contributing](#contributing)
- [License](#license)

## Install

Download a release archive (Linux `amd64`, macOS `arm64`) from the
[Releases page](https://github.com/adamdost-0/hashctl/releases) and verify its checksum:

```bash
curl -fsSLO https://github.com/adamdost-0/hashctl/releases/latest/download/hashctl-<version>-linux-amd64.tar.gz
curl -fsSLO https://github.com/adamdost-0/hashctl/releases/latest/download/hashctl-<version>-linux-amd64.tar.gz.sha256
shasum -a 256 -c hashctl-<version>-linux-amd64.tar.gz.sha256
tar -xzf hashctl-<version>-linux-amd64.tar.gz
```

Or use the helper script:

```bash
HASHCTL_PLATFORM=linux-amd64 ./scripts/install-hashctl.sh
```

Or build from source:

```bash
HASHCTL_VERSION=dev ./scripts/build-hashctl.sh   # writes bin/hashctl
```

## Usage

Global flags must appear **before** the command.

```bash
# Liveness / readiness
hashctl --api-url https://hashengine.example.com health
hashctl --api-url https://hashengine.example.com health ready

# Create, inspect, list, and cancel jobs
hashctl --api-url https://hashengine.example.com job create \
  --source-account myaccount --source-container mycontainer --prefix data/
hashctl --api-url https://hashengine.example.com job get <job_id>
hashctl --api-url https://hashengine.example.com job list
hashctl --api-url https://hashengine.example.com job cancel <job_id>

# Build and retrieve the manifest
hashctl --api-url https://hashengine.example.com manifest build <job_id>
hashctl --api-url https://hashengine.example.com manifest get --output-file manifest.xml <job_id>

# Apply signatures
hashctl --api-url https://hashengine.example.com sign first <job_id> --key-name mykey
hashctl --api-url https://hashengine.example.com sign second <job_id> --key-name mykey

# End-to-end smoke flows
hashctl --api-url https://hashengine.example.com smoke single-job \
  --source-account myaccount --source-container mycontainer --prefix data/
```

Run `hashctl help` or `hashctl <command> --help` for full command documentation.

## Configuration

Settings are resolved in precedence order (later sources win):

1. Built-in defaults (`--output human`, `--timeout 30s`, `--poll-interval 5s`, `--poll-timeout 10m`)
2. Config file `config.json` (located via `--config`, `HASHCTL_CONFIG`, or the OS config dir; only `api_url` and `output` are read)
3. Environment variables
4. CLI flags

| Environment variable | Purpose |
|---|---|
| `HASH_ENGINE_API` | API base URL (equivalent to `--api-url`) |
| `HASH_ENGINE_BEARER_TOKEN` | Bearer token (prefer `--bearer-token-file`) |
| `HASHCTL_CONFIG` | Path to `config.json` |

The API URL must use **HTTPS** for any non-loopback host. The literal `--bearer-token`
flag is intentionally unsupported; provide tokens via `HASH_ENGINE_BEARER_TOKEN` or a
`--bearer-token-file` whose permissions are `chmod 600`.

## Output modes

`--output human` (default) renders summaries; `--output json` emits machine-parseable
JSON. Secrets (JWTs, SAS query parameters, `AccountKey=`, high-entropy tokens) are
redacted from all output and error messages.

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success |
| `2` | Usage error (bad args, config, unknown command) |
| `3` | Transport failure (network error) |
| `4` | API client error (HTTP 4xx) |
| `5` | API server error (HTTP 5xx) |
| `6` | Poll timeout (job did not reach the target state) |

## Security

`hashctl` handles bearer tokens and signing credentials. See [SECURITY.md](SECURITY.md)
for the vulnerability disclosure process and the credential-handling design.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for the build/validation gate, commit conventions,
and PR process. By participating you agree to the
[Code of Conduct](CODE_OF_CONDUCT.md).

## License

[MIT](LICENSE) © Adam Dost
