#!/bin/bash

# Run after test_node and run_local_oracle.
# bash ./scripts/oracle/test_oracle_contracts.sh

export CHAIN_ID="${CHAIN_ID:-local-1}"
export NODE="${NODE:-http://localhost:26657}"
export TX_FLAGS="--from juno1 --keyring-backend test --chain-id $CHAIN_ID --gas 10000000 --fees 25000ujuno --broadcast-mode block --node $NODE --output json --yes --home $HOME/.juno1/"

# upload the contract & get code id
echo "Uploading contract..."
UPLOAD=$(junod tx wasm store ./scripts/oracle/oracle_querier.wasm $TX_FLAGS | jq -r '.txhash') && echo $UPLOAD
CODE_ID=$(junod q tx $UPLOAD --node $NODE --output json | jq -r '.logs[0].events[] | select(.type == "store_code").attributes[] | select(.key == "code_id").value') && echo "Code Id: $CODE_ID"

# Get the transaction upload hash to query in the next step
TX_HASH=$(junod tx wasm instantiate $CODE_ID "{}" --label "ORACLE QUERIER" $TX_FLAGS --no-admin | jq -r '.txhash') && echo $TX_HASH

# Query the logs for the contract address
CONTRACT_ADDR=$(junod query tx $TX_HASH --output json --node $NODE | jq -r '.logs[0].events[0].attributes[0].value') && echo "Address: $CONTRACT_ADDR"

junod q wasm contract-state smart $CONTRACT_ADDR '{"exchange_rate": {"denom":"ujuno"}}' --node $NODE