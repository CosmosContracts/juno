#!/usr/bin/env sh
set -eo pipefail

buf dep update
buf generate --template buf.gen.gogo.yaml

cp -r ./github.com/CosmosContracts/juno/x/* x/
cp -r ./github.com/cosmos/gaia/x/* x/

rm -rf ./github.com
