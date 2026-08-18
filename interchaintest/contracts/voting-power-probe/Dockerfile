# syntax=docker/dockerfile:1
# linux/amd64 manifest for the official rust:1.96.0-bookworm image.
FROM rust:1.96.0-bookworm@sha256:c993d32d95cc146bd12c84d66f0b924a6a96f3988325f39c144f2f9893dea120 AS builder

ARG BINARYEN_VERSION=120
ARG BINARYEN_SHA256=ddb097af51d1bdb17300d986b0de7d97422f1933dedb4c9eda3510e0bf4076cc
ARG ARTIFACT_SHA256=a074bc275b8b38eebd79db1603645ae71759a679385da2ccd82622f39f1f57af

RUN rustc --version | grep -Fx 'rustc 1.96.0 (ac68faa20 2026-05-25)' \
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
    && wasm-opt --enable-bulk-memory -Oz \
      target/wasm32-unknown-unknown/release/voting_power_probe.wasm \
      -o /voting_power_probe.wasm \
    && echo "${ARTIFACT_SHA256}  /voting_power_probe.wasm" | sha256sum --check --strict

FROM scratch AS artifact
COPY --from=builder /voting_power_probe.wasm /voting_power_probe.wasm
