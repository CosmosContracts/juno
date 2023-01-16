#!/bin/bash

# sh ./scripts/hermes/start.sh

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

HERMES="hermes --config $FOLDER/$config"

# First, lets make sure our chains are running and are healthy.
# Exit if theres an error.
{ 
    set -e 
    $HERMES health-check || echo "Chains are not healthy! don't forget to run them"
    $HERMES config validate || echo "Something is wrong with the config!"
}

CHAIN_A="local-1"
CHAIN_B="local-2"

# Create keys, same as the test_node.sh 2nd account
$HERMES keys add --chain $CHAIN_A --key-name default --mnemonic-file $FOLDER/relayer-mnemonic || true
$HERMES keys add --chain $CHAIN_B --key-name default --mnemonic-file $FOLDER/relayer-mnemonic || true

$HERMES create connection --a-chain $CHAIN_A --b-chain $CHAIN_B;

$HERMES create channel --a-port transfer --b-port transfer --a-chain $CHAIN_A --a-connection connection-0

$HERMES start