#!/usr/bin/env sh
set -eo pipefail

buf generate --template ./proto/buf.gen.pulsar.yaml --output ./api
