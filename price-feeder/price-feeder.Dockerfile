# Fetch base packages
FROM golang:1.19-alpine AS builder
ENV PACKAGES make git libc-dev gcc linux-headers
RUN apk add --no-cache $PACKAGES
WORKDIR /src/app/
COPY . .
# Build the binary
RUN cd price-feeder && make install

FROM alpine:3.14
RUN apk add bash curl jq
COPY --from=builder /go/bin/price-feeder /usr/local/bin/
EXPOSE 7171
CMD ["price-feeder"]
STOPSIGNAL SIGTERM
