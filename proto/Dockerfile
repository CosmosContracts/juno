# syntax=docker/dockerfile:1.6@sha256:ac85f380a63b13dfcefa89046420e1781752bab202122f8f50032edf31be0021

# --------------------------------------------------------
# Build arguments
# --------------------------------------------------------

ARG GO_VERSION="1.25.10"
ARG ALPINE_VERSION="3.22"

# --------------------------------------------------------
# Base image with Go toolchain
# --------------------------------------------------------

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION}@sha256:26b4d7113039cd51356bd7930ecafd1031d2975dc3b6940ec8ed09457e17cf95 AS base

ENV GOTOOLCHAIN=go1.25.10 \
    PATH=/go/bin:/usr/local/go/bin:/usr/local/bin:$PATH

RUN apk add --no-cache \
    ca-certificates=20260611-r0 \
    curl=8.14.1-r3 \
    git=2.49.1-r0 \
    make=4.4.1-r3 \
    yq-go=4.47.2-r3

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
