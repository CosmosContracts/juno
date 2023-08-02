#!/bin/bash
# TODO: Temp

# store contract as user

CHAIN_ID="local-1"
HOME_DIR="$HOME/.juno1"

export KEY="juno1"
export KEY2="juno2"

# localhost 26657
export JUNOD_NODE="http://localhost:26657"

# SudoMsg of JunoEndBlocker
junod tx wasm store ./scripts/cwmodules_example.wasm --from $KEY --gas 5000000 --chain-id $CHAIN_ID --gas-prices 0.025ujuno --node $JUNOD_NODE --broadcast-mode sync --yes
CODE_ID=1

junod tx wasm instantiate "1" "{}" --from $KEY --gas 5000000 --chain-id $CHAIN_ID --gas-prices 0.025ujuno --node $JUNOD_NODE --broadcast-mode sync --yes --label="A" --no-admin
CONTRACT="juno14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9skjuwg8"

# create a prop for the contract to be in the end blocker
junod tx gov submit-proposal proposal.json --from $KEY --gas 5000000 --chain-id $CHAIN_ID --gas-prices 0.025ujuno --node $JUNOD_NODE --broadcast-mode sync --yes

# junod q wasm contract-state smart juno14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9skjuwg8 '{"get_config":{}}'