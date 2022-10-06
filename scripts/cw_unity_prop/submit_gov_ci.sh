#!/bin/bash

if [ "$1" = "" ]
then
  echo "Usage: $0 1 arg required - juno address"
  exit
fi
sleep 10
CONTAINER_NAME="juno_node_1"
BINARY="docker exec -i $CONTAINER_NAME junod"
DENOM='ujunox'
CHAIN_ID='testing'
RPC='http://localhost:26657/'
TXFLAG="--gas-prices 0.1$DENOM --gas auto --gas-adjustment 1.3 -y -b block --chain-id $CHAIN_ID --node $RPC"
BLOCK_GAS_LIMIT=${GAS_LIMIT:-10000000} # mirrors mainnet

echo "Configured Block Gas Limit: $BLOCK_GAS_LIMIT"

# copy wasm to docker container
docker cp ./scripts/cw_unity_prop/cw_unity_prop.wasm $CONTAINER_NAME:/cw_unity_prop.wasm

# validator addr
VALIDATOR_ADDR=$($BINARY keys show validator --address)
echo "Validator address:"
echo $VALIDATOR_ADDR

BALANCE_1=$($BINARY q bank balances $VALIDATOR_ADDR)
echo "Pre-store balance:"
echo $BALANCE_1

echo "Address to deploy contracts: $1"
echo "TX Flags: $TXFLAG"

# errors from this point on are no bueno
set -e

# upload wasm
CONTRACT_CODE=$($BINARY tx wasm store "/cw_unity_prop.wasm" --from validator $TXFLAG --output json | jq -r '.logs[0].events[-1].attributes[0].value')
echo "Stored: $CONTRACT_CODE"

BALANCE_2=$($BINARY q bank balances $VALIDATOR_ADDR)
echo "Post-store balance:"
echo $BALANCE_2

# provision juno default user i.e. juno16g2rahf5846rxzp3fwlswy08fz8ccuwk03k57y
echo "clip hire initial neck maid actor venue client foam budget lock catalog sweet steak waste crater broccoli pipe steak sister coyote moment obvious choose" | $BINARY keys add test-user --recover --keyring-backend test

# instantiate
INIT='{
  "native_denom": "'"$DENOM"'",
  "withdraw_address": "'"$1"'",
  "withdraw_delay_in_days": 28
}'
echo "$INIT" | jq .

# --no-admin sent in test
$BINARY tx wasm instantiate $CONTRACT_CODE "$INIT" --from validator --label "juno unity prop" $TXFLAG --no-admin 
RES=$?

# get contract addr
CONTRACT_ADDRESS=$($BINARY q wasm list-contract-by-code $CONTRACT_CODE --output json | jq -r '.contracts[-1]')

# send contract funds
$BINARY tx bank send juno16g2rahf5846rxzp3fwlswy08fz8ccuwk03k57y $CONTRACT_ADDRESS 2000000ujunox $TXFLAG

$BINARY tx gov submit-proposal sudo-contract $CONTRACT_ADDRESS '{"execute_send": {"amount": "1000000", "recipient": "juno16g2rahf5846rxzp3fwlswy08fz8ccuwk03k57y"}}' \
  --from test-user $TXFLAG \
  --title "Prop title" \
  --description "LFG" \
  --deposit 500000000ujunox

$BINARY q gov proposal 1

# Print out config variables
printf "\n ------------------------ \n"
printf "Contract Variables \n\n"

echo "CODE_ID=$CONTRACT_CODE"
echo "CONTRACT_ADDRESS=$CONTRACT_ADDRESS"

echo $RES
exit $RES
