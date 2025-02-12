# docker build . -t cosmoscontracts/juno:latest
# docker run --rm -it cosmoscontracts/juno:latest /bin/sh
FROM golang:1.23-alpine AS go-builder

# this comes from standard alpine nightly file
#  https://github.com/rust-lang/docker-rust-nightly/blob/master/alpine3.12/Dockerfile
# with some changes to support our toolchain, etc
SHELL ["/bin/sh", "-ecuxo", "pipefail"]
# we probably want to default to latest and error
# since this is predominantly for dev use
# hadolint ignore=DL3018
RUN apk add --no-cache ca-certificates build-base git
# NOTE: add these to run with LEDGER_ENABLED=true
# RUN apk add libusb-dev linux-headers

WORKDIR /code

# Download dependencies and CosmWasm libwasmvm if found.
ADD go.mod go.sum ./
RUN set -eux; \
  export ARCH=$(uname -m); \
  WASM_VERSION=$(go list -m all | grep github.com/CosmWasm/wasmvm | awk '{print $2}'); \
  if [ ! -z "${WASM_VERSION}" ]; then \
  wget -O /lib/libwasmvm_muslc.a https://github.com/CosmWasm/wasmvm/releases/download/${WASM_VERSION}/libwasmvm_muslc.${ARCH}.a; \
  fi; \
  go mod download; \
  LIB_DIR="/lib" \
  if [ -e "$LIB_DIR/libwasmvm_muslc.x86_64.a" ]; then \
    ln -sf "$LIB_DIR/libwasmvm_muslc.x86_64.a" "$LIB_DIR/libwasmvm.x86_64.a" \
    echo "Created symlink: libwasmvm.x86_64.a -> libwasmvm_muslc.x86_64.a" \
  else \
    echo "Source file libwasmvm_muslc.x86_64.a not found, skipping." \
  fi \
  # Symlink for aarch64
  if [ -e "$LIB_DIR/libwasmvm_muslc.aarch64.a" ]; then \
    ln -sf "$LIB_DIR/libwasmvm_muslc.aarch64.a" "$LIB_DIR/libwasmvm.aarch64.a" \
    echo "Created symlink: libwasmvm.aarch64.a -> libwasmvm_muslc.aarch64.a" \
  else \
    echo "Source file libwasmvm_muslc.aarch64.a not found, skipping." \
  fi;
# Copy over code
COPY . /code/

# force it to use static lib (from above) not standard libgo_cosmwasm.so file
# then log output of file /code/bin/junod
# then ensure static linking
RUN LEDGER_ENABLED=false BUILD_TAGS=muslc LINK_STATICALLY=true make build \
  && file /code/bin/junod \
  && echo "Ensuring binary is statically linked ..." \
  && (file /code/bin/junod | grep "statically linked")

# --------------------------------------------------------

FROM alpine:3.21

COPY --from=go-builder /code/bin/junod /usr/bin/junod

COPY docker/* /opt/
RUN chmod +x /opt/*.sh

WORKDIR /opt

# rest server, comet p2p, comet rpc
EXPOSE 1317 26656 26657

CMD ["/usr/bin/junod", "version"]
