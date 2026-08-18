# syntax=docker/dockerfile:1
# linux/amd64 manifest for the official rust:1.85.1-bookworm image.
FROM rust:1.85.1-bookworm@sha256:bf7d87666c4da6eace19e06d21bc4859c6e2a5c97a21ac273b0e082112753cf0 AS builder

ARG BINARYEN_VERSION=120
ARG BINARYEN_SHA256=ddb097af51d1bdb17300d986b0de7d97422f1933dedb4c9eda3510e0bf4076cc
ARG ARTIFACT_SHA256=124d427ff478ec1d026f5412b72892db76a961b066989a299f66bcfb13d0b11a

RUN rustc --version | grep -Fx 'rustc 1.85.1 (4eb161250 2025-03-15)' \
    && rustup target add wasm32-unknown-unknown \
    && curl -fsSL \
      "https://github.com/WebAssembly/binaryen/releases/download/version_${BINARYEN_VERSION}/binaryen-version_${BINARYEN_VERSION}-x86_64-linux.tar.gz" \
      -o /tmp/binaryen.tar.gz \
    && echo "${BINARYEN_SHA256}  /tmp/binaryen.tar.gz" | sha256sum --check --strict \
    && tar -xzf /tmp/binaryen.tar.gz -C /opt \
    && ln -s "/opt/binaryen-version_${BINARYEN_VERSION}/bin/wasm-opt" /usr/local/bin/wasm-opt \
    && wasm-opt --version | grep -Fx "wasm-opt version ${BINARYEN_VERSION} (version_${BINARYEN_VERSION})" \
    && rm /tmp/binaryen.tar.gz

WORKDIR /build
COPY Cargo.toml Cargo.lock ./
COPY .cargo .cargo
COPY src src

RUN cargo test --locked \
    && cargo build --release --locked --target wasm32-unknown-unknown \
    && wasm-opt --disable-bulk-memory -Oz \
      target/wasm32-unknown-unknown/release/voting_power_probe.wasm \
      -o /voting_power_probe.wasm \
    && echo "${ARTIFACT_SHA256}  /voting_power_probe.wasm" | sha256sum --check --strict

FROM scratch AS artifact
COPY --from=builder /voting_power_probe.wasm /voting_power_probe.wasm
