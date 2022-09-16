# docker build . -t cosmoscontracts/juno:latest
# docker run --rm -it cosmoscontracts/juno:latest /bin/sh
FROM golang:1.18-alpine3.15 AS go-builder

# this comes from standard alpine nightly file
#  https://github.com/rust-lang/docker-rust-nightly/blob/master/alpine3.12/Dockerfile
# with some changes to support our toolchain, etc
SHELL ["/bin/ash", "-eo", "pipefail", "-c"]
# we probably want to default to latest and error
# since this is predominantly for dev use
# hadolint ignore=DL3018
RUN set -eux; apk add --no-cache ca-certificates build-base;

# hadolint ignore=DL3018
RUN apk add git
# NOTE: add these to run with LEDGER_ENABLED=true
# RUN apk add libusb-dev linux-headers

WORKDIR /code
COPY . /code/

# See https://github.com/CosmWasm/wasmvm/releases
ADD https://github.com/CosmWasm/wasmvm/releases/download/v1.1.0/libwasmvm_muslc.aarch64.a /lib/libwasmvm_muslc.aarch64.a
ADD https://github.com/CosmWasm/wasmvm/releases/download/v1.1.0/libwasmvm_muslc.x86_64.a /lib/libwasmvm_muslc.x86_64.a
RUN sha256sum /lib/libwasmvm_muslc.aarch64.a | grep 728993b91b35037ae8d9933c3a9ee018e49a7926571ce4109f55d9954efcbe9a
RUN sha256sum /lib/libwasmvm_muslc.x86_64.a | grep d06607db7bda6d3981f0717133584dd5480a6bca7b1e208b4526e68f3ccf3b31

# Copy the library you want to the final location that will be found by the linker flag `-lwasmvm_muslc`
RUN cp "/lib/libwasmvm_muslc.$(uname -m).a" /lib/libwasmvm_muslc.a

# force it to use static lib (from above) not standard libgo_cosmwasm.so file
# then log output of file /code/bin/junod
# then ensure static linking
RUN LEDGER_ENABLED=false BUILD_TAGS=muslc LINK_STATICALLY=true make build \
  && file /code/bin/junod \
  && echo "Ensuring binary is statically linked ..." \
  && (file /code/bin/junod | grep "statically linked")

# --------------------------------------------------------
FROM alpine:3.15

COPY --from=go-builder /code/bin/junod /usr/bin/junod

COPY docker/* /opt/
RUN chmod +x /opt/*.sh

WORKDIR /opt

# rest server
EXPOSE 1317
# tendermint p2p
EXPOSE 26656
# tendermint rpc
EXPOSE 26657

CMD ["/usr/bin/junod", "version"]