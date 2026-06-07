# Development container — host-agnostic Go 1.25 sandbox

`hashctl`'s `go.mod` requires **Go 1.25.0**. This dev container lets you build and
test the project in a sandbox that is agnostic to the host: the only host
requirement is a container runtime (Docker). No Go toolchain is installed on the
host, and the sandbox mirrors CI and the project's offline, stdlib-only policy
(`CGO_ENABLED=0`, `GOPROXY=off`, `GOSUMDB=off`).

## Two ways to use it

### 1. VS Code / Dev Containers

Open the repository in VS Code with the **Dev Containers** extension and choose
**Reopen in Container**. The container is built from
[`Dockerfile`](./Dockerfile) (`FROM golang:1.25-bookworm`, non-root `vscode`
user); `postCreateCommand` runs `go build ./...` and `go test ./...` to verify
the toolchain.

### 2. CLI helper — `scripts/sandbox.sh`

For terminal-only or CI-like use (no editor), run the validation gate or any
command inside the same image:

```bash
# Validation gate: go vet, go test, gofmt, go build
scripts/sandbox.sh

# Run an arbitrary command in the sandbox
scripts/sandbox.sh go test ./internal/hashctl -run TestRedact -v
```

The helper bind-mounts the repo, runs as the host user (so build artifacts are
not root-owned), and keeps the Go build cache inside the container. Override the
image with `HASHCTL_SANDBOX_IMAGE` if needed.

> Why a container instead of installing Go on the host? It guarantees the exact
> required toolchain (1.25.x) everywhere — laptops, CI, and agents — without
> mutating the host, and it keeps the offline/stdlib-only invariants intact.
