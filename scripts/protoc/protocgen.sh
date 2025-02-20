#!/usr/bin/env bash

## Install:
## + latest buf (v1.0.0-rc11 or later)

# docker build --pull --rm -f "contrib/devtools/Dockerfile" -t cosmossdk-proto:latest "contrib/devtools"
# docker run --rm -v $(pwd):/workspace --workdir /workspace cosmossdk-proto sh ./scripts/protocgen.sh

set -eo pipefail

cd proto
buf dep update
buf generate --template buf.gen.gogo.yaml
cd ..

# move proto files to the right places
cp -r ./github.com/CosmosContracts/juno/x/* x/
cp -r ./github.com/cosmos/gaia/x/* x/

rm -rf ./github.com