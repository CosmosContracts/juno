#!/usr/bin/env sh
set -eo pipefail

rm -rf gen/gogo
go tool buf generate --template ./proto/buf.gen.gogo.yaml

generated_root=./gen/gogo/github.com/CosmosContracts/juno
if [ -d "$generated_root/v31/x" ]; then
  generated_types="$generated_root/v31/x"
elif [ -d "$generated_root/x" ]; then
  generated_types="$generated_root/x"
else
  echo "generated Gogo types were not found under $generated_root" >&2
  exit 1
fi

cp -r "$generated_types"/. ./x/ || exit 1
rm -rf gen/gogo
