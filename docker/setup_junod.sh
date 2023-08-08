#!/bin/sh
#set -o errexit -o nounset -o pipefail

PASSWORD=${PASSWORD:-1234567890}
STAKE=${STAKE_TOKEN:-ustake}
FEE=${FEE_TOKEN:-ucosm}
CHAIN_ID=${CHAIN_ID:-testing}
MONIKER=${MONIKER:-node001}
KEYRING="--keyring-backend test"
TIMEOUT_COMMIT=${TIMEOUT_COMMIT:-5s}
BLOCK_GAS_LIMIT=${GAS_LIMIT:-10000000} # should mirror mainnet

echo "Configured Block Gas Limit: $BLOCK_GAS_LIMIT"

# check the genesis file
GENESIS_FILE="$HOME"/.juno/config/genesis.json
if [ -f "$GENESIS_FILE" ]; then
  echo "$GENESIS_FILE exists..."
else
  echo "$GENESIS_FILE does not exist. Generating..."

  junod init --chain-id "$CHAIN_ID" "$MONIKER"
  junod add-ica-config
  # staking/governance token is hardcoded in config, change this
  sed -i "s/\"stake\"/\"$STAKE\"/" "$GENESIS_FILE"
  # this is essential for sub-1s block times (or header times go crazy)
  sed -i 's/"time_iota_ms": "1000"/"time_iota_ms": "10"/' "$GENESIS_FILE"
  # change gas limit to mainnet value
  sed -i 's/"max_gas": "-1"/"max_gas": "'"$BLOCK_GAS_LIMIT"'"/' "$GENESIS_FILE"
  # change default keyring-backend to test
  sed -i 's/keyring-backend = "os"/keyring-backend = "test"/' "$HOME"/.juno/config/client.toml
fi

APP_TOML_CONFIG="$HOME"/.juno/config/app.toml
APP_TOML_CONFIG_NEW="$HOME"/.juno/config/app_new.toml
CONFIG_TOML_CONFIG="$HOME"/.juno/config/config.toml
if [ -n "$UNSAFE_CORS" ]; then
  echo "Unsafe CORS set... updating app.toml and config.toml"
  # sorry about this bit, but toml is rubbish for structural editing
  sed -n '1h;1!H;${g;s/# Enable defines if the API server should be enabled.\nenable = false/enable = true/;p;}' "$APP_TOML_CONFIG" > "$APP_TOML_CONFIG_NEW"
  mv "$APP_TOML_CONFIG_NEW" "$APP_TOML_CONFIG"
  # ...and breathe
  sed -i "s/enabled-unsafe-cors = false/enabled-unsafe-cors = true/" "$APP_TOML_CONFIG"
  sed -i "s/cors_allowed_origins = \[\]/cors_allowed_origins = \[\"\*\"\]/" "$CONFIG_TOML_CONFIG"
fi

# speed up block times for testing environments
sed -i "s/timeout_commit = \"5s\"/timeout_commit = \"$TIMEOUT_COMMIT\"/" "$CONFIG_TOML_CONFIG"

# are we running for the first time?
if ! junod keys show validator $KEYRING; then
  (echo "$PASSWORD"; echo "$PASSWORD") | junod keys add validator $KEYRING

  # hardcode the validator account for this instance
  echo "$PASSWORD" | junod genesis add-genesis-account validator "1000000000$STAKE,1000000000$FEE" $KEYRING

  # (optionally) add a few more genesis accounts
  for addr in "$@"; do
    echo $addr
    junod genesis add-genesis-account "$addr" "1000000000$STAKE,1000000000$FEE"
  done

  # submit a genesis validator tx
  ## Workraround for https://github.com/cosmos/cosmos-sdk/issues/8251
  (echo "$PASSWORD"; echo "$PASSWORD"; echo "$PASSWORD") | junod genesis gentx validator "250000000$STAKE" --chain-id="$CHAIN_ID" --amount="250000000$STAKE" $KEYRING
  ## should be:
  # (echo "$PASSWORD"; echo "$PASSWORD"; echo "$PASSWORD") | junod genesis gentx validator "250000000$STAKE" --chain-id="$CHAIN_ID"
  junod genesis collect-gentxs
fi
