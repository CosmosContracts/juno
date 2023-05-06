#!/usr/bin/env bash

# Run from the project root directory
# This script generates the swagger & openapi.yaml documentation for the rest API on port 1317
#
# Install the following::
# sudo npm install -g swagger2openapi swagger-merger swagger-combine
# go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@v1.16.0

# change to the scripts folder
cd "$(dirname `realpath "$0"`)"
# change to the root folder
cd ../

mkdir -p ./tmp-swagger-gen

# Get the paths used repos from go/pkg/mod
cosmos_sdk_dir=$(go list -f '{{ .Dir }}' -m github.com/cosmos/cosmos-sdk)
wasmd=$(go list -f '{{ .Dir }}' -m github.com/CosmWasm/wasmd)
gaia=$(go list -f '{{ .Dir }}' -m github.com/cosmos/gaia/v9)
ica=$(go list -f '{{ .Dir }}' -m github.com/cosmos/interchain-accounts)
pfm=$(go list -f '{{ .Dir }}' -m github.com/strangelove-ventures/packet-forward-middleware/v4)

proto_dirs=$(find ./proto "$cosmos_sdk_dir"/proto "$wasmd"/proto "$gaia"/proto "$ica"/proto "$pfm"/proto -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do

  # generate swagger files (filter query files)
  query_file=$(find "${dir}" -maxdepth 1 \( -name 'query.proto' -o -name 'service.proto' \))
  if [[ ! -z "$query_file" ]]; then
    protoc  \
    -I "proto" \
    -I "$cosmos_sdk_dir/third_party/proto" \
    -I "$cosmos_sdk_dir/proto" \
    -I "$wasmd/proto" \
    -I "$gaia/proto" \
    -I "$ica/proto" \
    -I "$pfm/proto" \
     "$query_file" \
    --swagger_out ./tmp-swagger-gen \
    --swagger_opt logtostderr=true \
    --swagger_opt fqn_for_swagger_name=true \
    --swagger_opt simple_operation_ids=true
  fi
done

# delete cosmos/mint path since juno uses its own module
rm -rf ./tmp-swagger-gen/cosmos/mint

# Fix circular definition in cosmos/tx/v1beta1/service.swagger.json
jq 'del(.definitions["cosmos.tx.v1beta1.ModeInfo.Multi"].properties.mode_infos.items["$ref"])' ./tmp-swagger-gen/cosmos/tx/v1beta1/service.swagger.json > ./tmp-swagger-gen/cosmos/tx/v1beta1/fixed-service.swagger.json

# Tag everything as "gRPC Gateway API"
perl -i -pe 's/"(Query|Service)"/"gRPC Gateway API"/' $(find ./tmp-swagger-gen -name '*.swagger.json' -print0 | xargs -0)

# Convert all *.swagger.json files into a single folder _all
files=$(find ./tmp-swagger-gen -name '*.swagger.json' -print0 | xargs -0)
mkdir -p ./tmp-swagger-gen/_all
counter=0
for f in $files; do
  echo "[+] $f"

  # check gaia first before cosmos
  if [[ "$f" =~ "gaia" ]]; then
    cp $f ./tmp-swagger-gen/_all/gaia-$counter.json
  elif [[ "$f" =~ "router" ]]; then
    cp $f ./tmp-swagger-gen/_all/pfm-$counter.json
  elif [[ "$f" =~ "cosmwasm" ]]; then
    cp $f ./tmp-swagger-gen/_all/cosmwasm-$counter.json
  elif [[ "$f" =~ "osmosis" ]]; then
    cp $f ./tmp-swagger-gen/_all/osmosis-$counter.json
  elif [[ "$f" =~ "juno" ]]; then
    cp $f ./tmp-swagger-gen/_all/juno-$counter.json
  elif [[ "$f" =~ "cosmos" ]]; then
    cp $f ./tmp-swagger-gen/_all/cosmos-$counter.json
  # elif [[ "$f" =~ "intertx" ]]; then
  #   cp $f ./tmp-swagger-gen/_all/intertx-$counter.json
  else
    cp $f ./tmp-swagger-gen/_all/other-$counter.json
  fi
  ((counter++))
done

# merges all the above into FINAL.json
python3 ./scripts/merge_protoc.py

# Makes a swagger temp file with reference pointers
swagger-combine ./tmp-swagger-gen/_all/FINAL.json -o ./docs/_tmp_swagger.yaml -f yaml --continueOnConflictingPaths --includeDefinitions

# extends out the *ref instances to their full value
swagger-merger --input ./docs/_tmp_swagger.yaml -o ./docs/swagger.yaml

# Derive openapi from swagger docs
swagger2openapi --patch ./docs/swagger.yaml --outfile ./docs/static/openapi.yml --yaml  

# clean swagger tmp files
rm ./docs/_tmp_swagger.yaml
rm -rf ./tmp-swagger-gen