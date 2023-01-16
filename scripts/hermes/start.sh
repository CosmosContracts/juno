#!/bin/bash

# sh ./scripts/hermes/start.sh

# CLI Args
CHAIN_A=${CHAIN_A:-"local-1"}
CHAIN_B=${CHAIN_B:-"local-2"}

A_PORT=${A_PORT:-transfer}
B_PORT=${B_PORT:-transfer}

CHANNEL_VERSION=${CHANNEL_VERSION:-ics20-1}



# Allows starting hermes from root dir of the project
FOLDER="$(dirname "$0")"

while [[ "$#" -gt 0 ]]; do
    case $1 in
        -c|--config) config="$2"; shift ;;
        *) echo "Unknown parameter passed: $1"; exit 1 ;;
    esac
    shift
done

# Verify we got the docker, either wasmd or osmosis
if [ -z "$config" ]; then config="config.toml"; fi

export HERMES="hermes --config $FOLDER/$config"
# export HERMES="hermes --config ./scripts/hermes/config.toml"

# First, lets make sure our chains are running and are healthy.
# Exit if theres an error.
{ 
    set -e 
    $HERMES health-check || echo "Chains are not healthy! don't forget to run them"
    $HERMES config validate || echo "Something is wrong with the config!"
}


# Create keys, same as the test_node.sh 2nd account
$HERMES keys add --chain $CHAIN_A --key-name default --mnemonic-file $FOLDER/relayer-mnemonic || true
$HERMES keys add --chain $CHAIN_B --key-name default --mnemonic-file $FOLDER/relayer-mnemonic || true

# Connections both ways on startup
$HERMES create connection --a-chain $CHAIN_A --b-chain $CHAIN_B;
$HERMES create connection --a-chain $CHAIN_B --b-chain $CHAIN_A;

$HERMES create channel --a-chain $CHAIN_A --a-port $A_PORT --b-chain $CHAIN_B --b-port $B_PORT --new-client-connection --channel-version $CHANNEL_VERSION
# channels both ways if both are transfer
if [ "$A_PORT" = "transfer" ] && [ "$B_PORT" = "transfer" ]; then
    $HERMES create channel --a-chain $CHAIN_B --a-port $B_PORT --b-chain $CHAIN_A --b-port $A_PORT --new-client-connection --channel-version $CHANNEL_VERSION
fi



$HERMES start