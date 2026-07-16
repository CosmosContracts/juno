#!/bin/sh
#set -o errexit -o nounset -o pipefail

BINARY=${BINARY:-"junod"}
DENOM=${DENOM:-"ujunox"}
CHAIN_ID=${CHAIN_ID:-"testchain-1"}
CHAIN_DIR=${CHAIN_DIR:-"$HOME/.juno"}
MONIKER=${MONIKER:-"node001"}
TIMEOUT_COMMIT=${TIMEOUT_COMMIT:-"2s"}
BLOCK_GAS_LIMIT=${BLOCK_GAS_LIMIT:-"10000000"}
KEYNAME=${KEYNAME:-"validator"}

if pgrep -x "$BINARY" >/dev/null; then
    echo "Terminating previous $BINARY..."
    killall "$BINARY"
fi
rm -rf "$CHAIN_DIR" >/dev/null 2>&1

if ! mkdir -p "$CHAIN_DIR" 2>/dev/null; then
    echo "Failed to create chain directory."
    exit 1
fi

$BINARY config set client chain-id $CHAIN_ID
$BINARY config set client keyring-backend test
$BINARY config set client keyring-default-keyname $KEYNAME
$BINARY config set client output json

$BINARY init $MONIKER --default-denom $DENOM

sed_inplace() {
    if [ "$(uname)" = "Darwin" ]; then
        sed -i '' "$@"
    else
        sed -i "$@"
    fi
}
sed_inplace 's/"time_iota_ms": "1000"/"time_iota_ms": "10"/' "$CHAIN_DIR"/config/genesis.json
sed_inplace 's/"max_gas": "-1"/"max_gas": "'"$BLOCK_GAS_LIMIT"'"/' "$CHAIN_DIR"/config/genesis.json

$BINARY config set app api.enable true
$BINARY config set app api.swagger false
$BINARY config set app api.enabled-unsafe-cors true

$BINARY config set config rpc.cors_allowed_origins "*" --skip-validate
$BINARY config set config consensus.timeout_commit "$TIMEOUT_COMMIT" --skip-validate

$BINARY keys add $KEYNAME
$BINARY genesis add-ica-config
$BINARY genesis add-genesis-account $KEYNAME "1000000000$DENOM"

for addr in "$@"; do
  echo $addr
  $BINARY genesis add-genesis-account "$addr" "1000000000$DENOM" >/dev/null 2>&1
done

$BINARY genesis gentx $KEYNAME "250000000$DENOM"
$BINARY genesis collect-gentxs >/dev/null 2>&1
