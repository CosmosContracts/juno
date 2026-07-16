# syntax=docker/dockerfile:1.6

# --------------------------------------------------------
# Build arguments
# --------------------------------------------------------

ARG GO_VERSION="1.25.2"
ARG ALPINE_VERSION="3.22"

# --------------------------------------------------------
# Base image with Go toolchain
# --------------------------------------------------------

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS base

ENV GOTOOLCHAIN=go1.25.2 \
    PATH=/go/bin:/usr/local/go/bin:/usr/local/bin:$PATH

RUN apk add --no-cache \
    ca-certificates \
    curl \
    git \
    make \
    yq

# --------------------------------------------------------
# Tools stage warms Go module/tool caches
# --------------------------------------------------------

FROM base AS tools

WORKDIR /workspace
COPY go.mod go.sum ./

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/go/pkg/mod \
    go mod download

# --------------------------------------------------------
# Runner
# --------------------------------------------------------

FROM base

COPY --from=tools /go/pkg/mod /go/pkg/mod

RUN addgroup -S -g 1000 protobuild \
    && adduser -S -D -G protobuild -u 1000 protobuild \
    && mkdir -p /workspace \
    && chown -R protobuild:protobuild /go /workspace

WORKDIR /workspace

USER protobuild
