# syntax=docker/dockerfile:1.6

# --------------------------------------------------------
# Arguments
# --------------------------------------------------------

ARG PROTO_BUILDER_VERSION="0.17.1"

# --------------------------------------------------------
# Builder
# --------------------------------------------------------

FROM ghcr.io/cosmos/proto-builder:${PROTO_BUILDER_VERSION}
ENV GOTOOLCHAIN=go1.25.2

USER root

RUN apk add --no-cache yq
