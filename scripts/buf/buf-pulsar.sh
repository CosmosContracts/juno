#!/usr/bin/env sh
set -eo pipefail

go tool buf generate --template ./proto/buf.gen.pulsar.yaml --output ./api
