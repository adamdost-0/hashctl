# Host-agnostic Go sandbox for building and testing hashctl.
#
# Pinned to the major.minor required by go.mod (`go 1.25.0`). The official
# golang image tracks the latest 1.25.x patch, which satisfies the toolchain
# requirement without the host needing any Go installed.
FROM golang:1.25-bookworm

# hashctl is stdlib-only and builds offline (ADR-0001). Bake the offline,
# CGO-free invariants into the environment so in-container commands match CI.
ENV CGO_ENABLED=0 \
    GOPROXY=off \
    GOSUMDB=off

# A non-root user so files created in the bind-mounted workspace are owned by a
# normal user rather than root. VS Code remaps this UID to the host user via
# updateRemoteUserUID (default true); the CLI helper (scripts/sandbox.sh) sets
# --user explicitly and does not depend on this UID.
ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=$USER_UID
RUN apt-get update \
    && apt-get install -y --no-install-recommends git ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && groupadd --gid $USER_GID $USERNAME \
    && useradd --uid $USER_UID --gid $USER_GID -m $USERNAME

USER $USERNAME
