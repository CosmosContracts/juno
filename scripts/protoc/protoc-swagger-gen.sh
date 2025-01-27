#!/usr/bin/env sh

set -eo pipefail

# install the following npm packages globally
# sudo npm install -g swagger2openapi swagger-merger swagger-combine

# run go mod tidy to get the latest dependencies before running this script

prepare_swagger_gen() {
  go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@v1.16.0
}

prepare_swagger_gen
go mod tidy
mkdir -p tmp-swagger-gen

cosmos_sdk_dir=$(go list -f '{{ .Dir }}' -m github.com/cosmos/cosmos-sdk)
wasmd=$(go list -f '{{ .Dir }}' -m github.com/CosmWasm/wasmd)
pfm=$(go list -f '{{ .Dir }}' -m "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8")

cd proto
proto_dirs=$(find ./ "$cosmos_sdk_dir"/proto "$wasmd"/proto "$pfm"/proto -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  # generate swagger files (filter query files)
  query_file=$(find "${dir}" -maxdepth 1 \( -name 'query.proto' -o -name 'service.proto' \))

  if [ -n "$query_file" ]; then
    buf generate --template buf.gen.swagger.yaml $query_file
  fi
done

cd ..

# Fix circular definition in cosmos by removing them
jq 'del(.definitions["cosmos.tx.v1beta1.ModeInfo.Multi"].properties.mode_infos.items["$ref"])' ./tmp-swagger-gen/cosmos/tx/v1beta1/service.swagger.json > ./tmp-swagger-gen/cosmos/tx/v1beta1/fixed_service.swagger.json
jq 'del(.definitions["cosmos.autocli.v1.ServiceCommandDescriptor"].properties.sub_commands)' ./tmp-swagger-gen/cosmos/autocli/v1/query.swagger.json > ./tmp-swagger-gen/cosmos/autocli/v1/fixed_query.swagger.json

rm -rf ./tmp-swagger-gen/cosmos/tx/v1beta1/service.swagger.json
rm -rf ./tmp-swagger-gen/cosmos/autocli/v1/query.swagger.json

# delete cosmos/mint path since juno uses its own module
rm -rf tmp-swagger-gen/cosmos/mint

# Tag everything as "gRPC Gateway API"
find ./tmp-swagger-gen -name '*.swagger.json' -print0 | xargs -0 perl -i -pe 's/"(Query|Service)"/"gRPC Gateway API"/'

# Convert all *.swagger.json files into a single folder _all
files=$(find ./tmp-swagger-gen -name '*.swagger.json' -print0 | xargs -0)
mkdir -p ./tmp-swagger-gen/_all
counter=0
for f in $files; do
  echo "[+] $f"

  case "$f" in
    *router*) cp "$f" ./tmp-swagger-gen/_all/pfm-$counter.json ;;
    *cosmwasm*) cp "$f" ./tmp-swagger-gen/_all/cosmwasm-$counter.json ;;
    *osmosis*) cp "$f" ./tmp-swagger-gen/_all/osmosis-$counter.json ;;
    *juno*) cp "$f" ./tmp-swagger-gen/_all/juno-$counter.json ;;
    *cosmos*) cp "$f" ./tmp-swagger-gen/_all/cosmos-$counter.json ;;
    *) cp "$f" ./tmp-swagger-gen/_all/other-$counter.json ;;
  esac

  counter=$(expr $counter + 1)
done

# merges all the above into FINAL.json
python3 scripts/protoc/merge_protoc.py

# Makes a swagger temp file with reference pointers
swagger-combine ./tmp-swagger-gen/_all/FINAL.json -o ./tmp-swagger-gen/tmp_swagger.yaml -f yaml --continueOnConflictingPaths true --includeDefinitions true

# extends out the *ref instances to their full value
swagger-merger --input ./tmp-swagger-gen/tmp_swagger.yaml -o ./docs/static/swagger.yaml

rm -rf tmp-swagger-gen