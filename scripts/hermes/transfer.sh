#!/bin/sh

export JUNOD_NODE="http://localhost:26657"
CHAIN_A_ARGS="--from juno1 --keyring-backend test --chain-id local-1 --home $HOME/.juno1/ --node http://localhost:26657 --yes"

# junod q ibc channel channels

# Send from local-1 to local-2 via the relayer
junod tx ibc-transfer transfer transfer channel-0 juno1hj5fveer5cjtn4wd6wstzugjfdxzl0xps73ftl 9ujuno $CHAIN_A_ARGS --packet-timeout-height 0-0

sleep 6

# check the query on the other chain to ensure it went through
junod q bank balances juno1hj5fveer5cjtn4wd6wstzugjfdxzl0xps73ftl --chain-id local-2 --node http://localhost:36657