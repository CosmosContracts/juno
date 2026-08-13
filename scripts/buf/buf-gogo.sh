#!/usr/bin/env sh
set -eo pipefail

go tool buf generate --template ./proto/buf.gen.gogo.yaml

cp -r ./github.com/CosmosContracts/juno/x/* x/
rm -rf ./github.com
