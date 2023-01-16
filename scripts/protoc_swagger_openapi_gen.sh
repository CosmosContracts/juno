#!/usr/bin/env bash

# Run from the project root directory
# This script generates the swagger & openapi.yaml documentation for the rest API on port 1317

# change to the scripts folder
cd "$(dirname `realpath "$0"`)"
# change to the root folder
cd ../

mkdir -p ./tmp-swagger-gen

# Get the path of the cosmos-sdk repo from go/pkg/mod
cosmos_sdk_dir=$(go list -f '{{ .Dir }}' -m github.com/cosmos/cosmos-sdk)
wasmd=$(go list -f '{{ .Dir }}' -m github.com/CosmWasm/wasmd)
token_factory=$(go list -f '{{ .Dir }}' -m github.com/CosmWasm/token-factory)
gaia=$(go list -f '{{ .Dir }}' -m github.com/cosmos/gaia/v8)
ica=$(go list -f '{{ .Dir }}' -m github.com/cosmos/interchain-accounts)

proto_dirs=$(find ./proto "$cosmos_sdk_dir"/proto "$wasmd"/proto "$token_factory"/proto "$gaia"/proto "$ica"/proto -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do

  # generate swagger files (filter query files)
  query_file=$(find "${dir}" -maxdepth 1 \( -name 'query.proto' -o -name 'service.proto' \))
  if [[ ! -z "$query_file" ]]; then
    # 2. Get swagger protoc plugin with `go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@v1.16.0`

    protoc  \
    -I "proto" \
    -I "$cosmos_sdk_dir/third_party/proto" \
    -I "$cosmos_sdk_dir/proto" \
    -I "$wasmd/proto" \
    -I "$token_factory/proto" \
    -I "$gaia/proto" \
    -I "$ica/proto" \
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


# convert all *.swagger.json files into a single folder _all
files=$(find ./tmp-swagger-gen -name '*.swagger.json' -print0 | xargs -0)
mkdir -p ./tmp-swagger-gen/_all
counter=0
for f in $files; do
  echo "$f"

  # if juno in f, then do something
  if [[ "$f" =~ "gaia" ]]; then
    cp $f ./tmp-swagger-gen/_all/gaia-$counter.json
  elif [[ "$f" =~ "cosmwasm" ]]; then
    cp $f ./tmp-swagger-gen/_all/cosmwasm-$counter.json
  elif [[ "$f" =~ "osmosis" ]]; then
    cp $f ./tmp-swagger-gen/_all/osmosis-$counter.json
  elif [[ "$f" =~ "juno" ]]; then
    cp $f ./tmp-swagger-gen/_all/juno-$counter.json
  elif [[ "$f" =~ "cosmos" ]]; then
    cp $f ./tmp-swagger-gen/_all/cosmos-$counter.json
  elif [[ "$f" =~ "tokenfactory" ]]; then
    cp $f ./tmp-swagger-gen/_all/tokenfactory-$counter.json
  elif [[ "$f" =~ "intertx" ]]; then
    cp $f ./tmp-swagger-gen/_all/intertx-$counter.json
  else 
    echo "$f"
    cp $f ../tmp-swagger-gen/_all/cosmos-$counter.json
  fi
  ((counter++))
done

# merges all the above into FINAL.json
python3 ./scripts/merge_protoc.py

# Makes a swagger temp file with reference pointers
swagger-combine ./tmp-swagger-gen/_all/FINAL.json -o ./docs/_tmp_swagger.yaml -f yaml --continueOnConflictingPaths --includeDefinitions
# swagger-merge -v "v12" -t "Juno Network" -d "A REST interface for Juno's queries and transactions" -s https,http -p "/" ./docs/_tmp_swagger.yaml -o swagger.json

# extends it out the ref pointers
swagger-merger --input ./docs/_tmp_swagger.yaml -o ./docs/swagger.yaml

# Derive openapi & swagger from config.json
swagger2openapi --patch ./docs/swagger.yaml --outfile ./docs/static/openapi.yml --yaml  

# clean swagger tmp files
rm ./docs/_tmp_swagger.yaml
rm -rf ./tmp-swagger-gen