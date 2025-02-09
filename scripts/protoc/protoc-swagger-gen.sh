#!/usr/bin/env bash
set -euo pipefail

prepare_swagger_gen() {
  go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@v1.16.0
}

run_npm() {
  if command -v npm &>/dev/null; then
    npm "$@"
  else
    if command -v docker &>/dev/null; then
      docker run --rm -v "$project_root":/workspace -w /workspace node:14-alpine npm "$@"
    else
      echo "Error: Neither npm nor docker is available." >&2
      exit 1
    fi
  fi
}

run_swagger_combine() {
  if command -v swagger-combine &>/dev/null; then
    swagger-combine "$@"
  else
    if command -v npm &>/dev/null; then
      npx swagger-combine "$@"
    else
      docker run --rm -v "$project_root":/workspace -w /workspace node:14-alpine npx swagger-combine "$@"
    fi
  fi
}

run_swagger_merger() {
  if command -v swagger-merger &>/dev/null; then
    swagger-merger "$@"
  else
    if command -v npm &>/dev/null; then
      npx swagger-merger "$@"
    else
      docker run --rm -v "$project_root":/workspace -w /workspace node:14-alpine npx swagger-merger "$@"
    fi
  fi
}

go mod tidy
prepare_swagger_gen
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

# Fix circular definitions in cosmos by removing them
jq 'del(.definitions["cosmos.tx.v1beta1.ModeInfo.Multi"].properties.mode_infos.items["$ref"])' ./tmp-swagger-gen/cosmos/tx/v1beta1/service.swagger.json > ./tmp-swagger-gen/cosmos/tx/v1beta1/fixed_service.swagger.json
jq 'del(.definitions["cosmos.autocli.v1.ServiceCommandDescriptor"].properties.sub_commands)' ./tmp-swagger-gen/cosmos/autocli/v1/query.swagger.json > ./tmp-swagger-gen/cosmos/autocli/v1/fixed_query.swagger.json

rm -rf ./tmp-swagger-gen/cosmos/tx/v1beta1/service.swagger.json
rm -rf ./tmp-swagger-gen/cosmos/autocli/v1/query.swagger.json

# Delete cosmos/mint path since juno uses its own module
rm -rf ./tmp-swagger-gen/cosmos/mint

mkdir -p ./tmp-swagger-gen/_all

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

# Ensure jq is available.
if ! command -v jq &> /dev/null; then
  echo "Error: jq is not installed. Please install jq." >&2
  exit 1
fi

# Determine directories.
current_dir="$(dirname "$(realpath "$0")")"
project_root="$(dirname "$(dirname "$current_dir")")"
all_dir="$project_root/tmp-swagger-gen/_all"

# Extract the version from go.mod.
version=$(grep "^module" "$project_root/go.mod" | head -n1 | awk -F'/' '{print $NF}' | tr -d ' ')
if [ -z "$version" ]; then
  echo "Could not find version in go.mod" >&2
  exit 1
fi

# Build the base JSON structure.
base_json=$(jq -n --arg version "$version" '{
  swagger: "2.0",
  info: { title: "Juno network", version: $version },
  consumes: ["application/json"],
  produces: ["application/json"],
  paths: {},
  definitions: {}
}')

# Save the base JSON to a temporary file.
temp_file=$(mktemp)
echo "$base_json" > "$temp_file"

# Loop through all JSON files in the target directory and merge their "paths" and "definitions".
for file in "$all_dir"/*.json; do
  # Skip FINAL.json to avoid merging our final output.
  if [[ $(basename "$file") == "FINAL.json" ]]; then
    continue
  fi
  new_json=$(cat "$file")
  temp_file2=$(mktemp)
  jq --argjson new "$new_json" '
    .paths += ($new.paths // {}) |
    .definitions += ($new.definitions // {})' "$temp_file" > "$temp_file2"
  mv "$temp_file2" "$temp_file"
done

# Loop through all paths and methods to update any "operationId" by appending a random 5-character suffix.
while read -r path method; do
  # The subshell disables pipefail to avoid SIGPIPE errors from head.
  suffix=$( (set +o pipefail; tr -dc 'A-Z0-9' </dev/urandom | head -c 5) )
  temp_file2=$(mktemp)
  jq --arg path "$path" --arg method "$method" --arg suffix "$suffix" '
    if (.paths[$path][$method] | has("operationId"))
    then .paths[$path][$method].operationId |= (. + "_" + $suffix)
    else . end' "$temp_file" > "$temp_file2"
  mv "$temp_file2" "$temp_file"
done < <(jq -r '.paths | to_entries[] | "\(.key) \(.value | keys[])"' "$temp_file")

# Save the final merged JSON to FINAL.json.
jq . "$temp_file" > "$all_dir/FINAL.json"
rm "$temp_file"

echo "Merged JSON saved to $all_dir/FINAL.json"

# Use swagger-combine to create a swagger temp file with reference pointers.
run_swagger_combine "$all_dir/FINAL.json" -o "./tmp-swagger-gen/tmp_swagger.yaml" -f yaml --continueOnConflictingPaths true --includeDefinitions true

# Use swagger-merger to extend the $ref instances to their full value.
run_swagger_merger --input "./tmp-swagger-gen/tmp_swagger.yaml" -o "./docs/static/swagger.yaml"

# Cleanup.
rm -rf tmp-swagger-gen

echo "Swagger generation complete. Output at ./docs/static/swagger.yaml"
