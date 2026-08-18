# syntax=docker/dockerfile:1.10
# The multi-architecture builder and Go toolchain are pinned by index digest.
FROM golang:1.25.10-alpine3.22@sha256:26b4d7113039cd51356bd7930ecafd1031d2975dc3b6940ec8ed09457e17cf95 AS builder
ARG APK_CA_CERTIFICATES=20260611-r0
ARG APK_BUILD_BASE=0.5-r3
ARG APK_LINUX_HEADERS=6.14.2-r0
ARG APK_GIT=2.49.1-r0
RUN apk add --no-cache \
    "ca-certificates=$APK_CA_CERTIFICATES" \
    "build-base=$APK_BUILD_BASE" \
    "linux-headers=$APK_LINUX_HEADERS" \
    "git=$APK_GIT" && \
    apk info -v | LC_ALL=C sort >/build-dependencies.txt
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
ARG VERSION
ARG COMMIT
ARG SOURCE_DATE_EPOCH
ENV GOTOOLCHAIN=go1.25.10 SOURCE_DATE_EPOCH=$SOURCE_DATE_EPOCH
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    LEDGER_ENABLED=false BUILD_TAGS=muslc LINK_STATICALLY=true \
    VERSION="$VERSION" COMMIT="$COMMIT" LDFLAGS="-buildid=" make build && \
    file bin/junod | grep -q 'statically linked' && \
    go version -m bin/junod > /junod.modules

FROM scratch AS artifact
COPY --from=builder /src/bin/junod /junod
COPY --from=builder /junod.modules /junod.modules
COPY --from=builder /build-dependencies.txt /build-dependencies.txt

FROM golang:1.25.10-alpine3.22@sha256:26b4d7113039cd51356bd7930ecafd1031d2975dc3b6940ec8ed09457e17cf95 AS image
ARG VERSION
ARG COMMIT
ARG SOURCE_REPOSITORY=https://github.com/CosmosContracts/juno
LABEL org.opencontainers.image.title="Juno" \
      org.opencontainers.image.version="$VERSION" \
      org.opencontainers.image.revision="$COMMIT" \
      org.opencontainers.image.source="$SOURCE_REPOSITORY"
COPY --from=builder /src/bin/junod /bin/junod
ENTRYPOINT ["/bin/junod"]