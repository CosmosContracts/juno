# docker build . -t cosmoscontracts/junod:latest
# docker run --rm -it cosmoscontracts/junod:latest /bin/sh
FROM golang:1.17-alpine3.13 AS go-builder

# this comes from standard alpine nightly file
#  https://github.com/rust-lang/docker-rust-nightly/blob/master/alpine3.12/Dockerfile
# with some changes to support our toolchain, etc
RUN set -eux; apk add --no-cache ca-certificates build-base;

RUN apk add git
# NOTE: add these to run with LEDGER_ENABLED=true
# RUN apk add libusb-dev linux-headers

WORKDIR /code
COPY . /code/

# See https://github.com/CosmWasm/wasmvm/releases
ADD https://github.com/CosmWasm/wasmvm/releases/download/v1.0.0-beta/libwasmvm_muslc.a /lib/libwasmvm_muslc.a
RUN sha256sum /lib/libwasmvm_muslc.a | grep 2ea10ad5e489b5ede1aa4061d4afa8b2ddd39718ba7b8689690b9c07a41d678e

# force it to use static lib (from above) not standard libgo_cosmwasm.so file
RUN LEDGER_ENABLED=false BUILD_TAGS=muslc make build

# --------------------------------------------------------
FROM alpine:3.13

WORKDIR /root

COPY --from=go-builder /code/bin/junod /usr/bin/junod

COPY scripts /scripts

ENTRYPOINT [ "/usr/bin/junod" ]

CMD [ "version" ]