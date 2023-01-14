#!/usr/bin/env bash

set -eo pipefail

# change to the scripts folder
cd "$(dirname `realpath "$0"`)"
# change to the root folder
cd ../

mkdir -p ./tmp-swagger-gen

# Get the path of the cosmos-sdk repo from go/pkg/mod
cosmos_sdk_dir=$(go list -f '{{ .Dir }}' -m github.com/cosmos/cosmos-sdk)

proto_dirs=$(find ./proto "$cosmos_sdk_dir"/proto -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do

  # generate swagger files (filter query files)
  query_file=$(find "${dir}" -maxdepth 1 \( -name 'query.proto' -o -name 'service.proto' \))
  if [[ ! -z "$query_file" ]]; then
    # 2. Get swagger protoc plugin with `go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@v1.16.0`

    protoc  \
    -I "proto" \
    -I "$cosmos_sdk_dir/third_party/proto" \
    -I "$cosmos_sdk_dir/proto" \
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

# (
#   cd ./client/docs

#   # Generate config.json
#   # There's some operationIds naming collision, for sake of automation we're
#   # giving all of them a unique name
#   find ../../tmp-swagger-gen -name 'query.swagger.json' -o -name 'fixed-service.swagger.json' | 
#     sort |
#     awk '{print "{\"url\":\""$1"\",\"operationIds\":{\"rename\":{\"Params\":\""$1"Params\",\"Pool\":\""$1"Pool\",\"DelegatorValidators\":\""$1"DelegatorValidators\",\"UpgradedConsensusState\":\""$1"UpgradedConsensusState\"}}}"}' |
#     jq -s '{swagger:"2.0","info":{"title":"Juno Network","description":"A REST interface for queries and transactions","version":"'"${CHAIN_VERSION}"'"},apis:.} | .apis += [{"url":"./swagger_legacy.yaml","dereference":{"circular":"ignore"}}]' > ./config.json

#   # Derive openapi & swagger from config.json
#   swagger-combine ./config.json -o static/swagger/swagger.yaml -f yaml --continueOnConflictingPaths --includeDefinitions
#   mkdir -p static/openapi && swagger2openapi --patch static/swagger/swagger.yaml --outfile static/openapi/openapi.yaml --yaml
#   redoc-cli build static/openapi/openapi.yaml --output ./static/openapi/index.html
# )

(
  cd ./docs

  # Generate config.json
  # There's some operationIds naming collision, for sake of automation we're
  # giving all of them a unique name
  find ../tmp-swagger-gen -name 'query.swagger.json' -o -name 'fixed-service.swagger.json' | 
    sort |
    awk '{print "{\"url\":\""$1"\",\"operationIds\":{\"rename\":{\"Params\":\""$1"Params\",\"Pool\":\""$1"Pool\",\"DelegatorValidators\":\""$1"DelegatorValidators\",\"UpgradedConsensusState\":\""$1"UpgradedConsensusState\"}}}"}' |
    jq -s '{swagger:"2.0","info":{"title":"Juno Network","description":"A REST interface for queries and transactions","version":"'"${CHAIN_VERSION}"'"},apis:.} | .apis += [{"url":"./swagger.yaml","dereference":{"circular":"ignore"}}]' > ./config.json

  # Derive openapi & swagger from config.json
  swagger-combine ./config.json -o swagger.yaml -f yaml --continueOnConflictingPaths --includeDefinitions
  swagger2openapi --patch swagger.yaml --outfile static/openapi.yml --yaml
  # redoc-cli build static/openapi.yaml --output ./static/openapi/index.html
)

# clean swagger tmp files
rm -rf ./tmp-swagger-gen