#!/bin/bash

# bash ./scripts/bindings/test_bindings.sh

export CHAIN_ID="${CHAIN_ID:-local-1}"
export JUNOD_NODE="${NODE:-http://localhost:26657}"
export TX_FLAGS="--from juno1 --keyring-backend test --chain-id $CHAIN_ID --gas 10000000 --fees 20000ujuno --broadcast-mode block --node $JUNOD_NODE --output json --yes --home $HOME/.juno1/"
JUNO_KEY_ADDR="juno1hj5fveer5cjtn4wd6wstzugjfdxzl0xps73ftl"

# upload the contract & get code id
echo "Uploading contract..."
UPLOAD=$(junod tx wasm store ./scripts/bindings/increment.wasm $TX_FLAGS | jq -r '.txhash') && echo $UPLOAD
CODE_ID=$(junod q tx $UPLOAD --node $JUNOD_NODE --output json | jq -r '.logs[0].events[] | select(.type == "store_code").attributes[] | select(.key == "code_id").value') && echo "Code Id: $CODE_ID"

# Get the transaction upload hash to query in the next step
TX_HASH=$(junod tx wasm instantiate $CODE_ID '{"count":0}' --label "bindings" $TX_FLAGS --no-admin | jq -r '.txhash') && echo $TX_HASH

# Query the logs for the contract address
CONTRACT_ADDR=$(junod query tx $TX_HASH --output json --node $JUNOD_NODE | jq -r '.logs[0].events[0].attributes[0].value') && echo "Address: $CONTRACT_ADDR"

# create proposal
junod tx gov submit-proposal software-upgrade vtest2 --upgrade-height 6999999 --title "test" --description "test" --deposit 200000ujuno $TX_FLAGS
ID="1" && junod tx gov deposit $ID 800000ujuno $TX_FLAGS && junod tx gov vote $ID yes $TX_FLAGS && junod q gov proposal $ID

PAYLOAD=$(printf '{"gov_vote":{"proposal_id":%s,"voter":"%s"}}' $ID $JUNO_KEY_ADDR)
junod q wasm contract-state smart $CONTRACT_ADDR "$PAYLOAD" --node $JUNOD_NODE

# data:
#   vote:
#     options:
#     - option: "yes"
#       weight: "1.000000000000000000"
#     proposal_id: 1
#     voter: juno1hj5fveer5cjtn4wd6wstzugjfdxzl0xps73ftl