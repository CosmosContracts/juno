# syntax=docker/dockerfile:1

# --------------------------------------------------------
# Arguments
# --------------------------------------------------------

ARG GO_VERSION="1.25.2"
ARG ALPINE_VERSION="3.22"

# --------------------------------------------------------
# Builder
# --------------------------------------------------------

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder
ENV GOTOOLCHAIN=go1.25.2

RUN apk add --no-cache \
    ca-certificates \
    build-base \
    linux-headers \
    git

WORKDIR /juno

# Copy Go dependencies
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/go/pkg/mod \
    go mod download

# Fetch wasmvm
RUN WASMVM_VERSION=$(go list -m github.com/CosmWasm/wasmvm/v2 | cut -d ' ' -f 2) && \
    wget https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION/libwasmvm_muslc.$(uname -m).a \
    -O /lib/libwasmvm_muslc.$(uname -m).a && \
    # verify checksum
    wget https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION/checksums.txt -O /tmp/checksums.txt && \
    sha256sum /lib/libwasmvm_muslc.$(uname -m).a | grep $(cat /tmp/checksums.txt | grep libwasmvm_muslc.$(uname -m) | cut -d ' ' -f 1)

# Copy source code
COPY . .

# Build binary
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/go/pkg/mod \
    LEDGER_ENABLED=false BUILD_TAGS=muslc LINK_STATICALLY=true make build \
    && file /juno/bin/junod \
    && echo "Ensuring binary is statically linked ..." \
    && (file /juno/bin/junod | grep "statically linked")

# --------------------------------------------------------
# Runner
# --------------------------------------------------------

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION}
COPY --from=builder /juno/bin/junod /bin/junod

ENV HOME=/.juno
WORKDIR $HOME

EXPOSE 26656
EXPOSE 26657
EXPOSE 1317
EXPOSE 9090

ENTRYPOINT ["junod"]