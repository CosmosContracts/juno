#!/bin/sh
# set -e 
# set -o pipefail

# Home path in the docker container
JUNO_HOME=/juno/.juno
CONFIG_FOLDER=$JUNO_HOME/config

# val - juno1jxa3ksucx7ter57xyuczvmk6qkeqmqvj37g237
DEFAULT_MNEMONIC="blame tube add leopard fire next exercise evoke young team payment senior know estate mandate negative actual aware slab drive celery elevator burden utility"
DEFAULT_CHAIN_ID="localjuno"
DEFAULT_MONIKER="val"

# Override default values with environment variables
MNEMONIC=${MNEMONIC:-$DEFAULT_MNEMONIC}
CHAIN_ID=${CHAIN_ID:-$DEFAULT_CHAIN_ID}
MONIKER=${MONIKER:-$DEFAULT_MONIKER}

install_prerequisites () {
    apk add -q --no-cache \
        dasel \
        python3 \
        py3-pip
}

edit_config () {
    # Remove seeds
    dasel put string -f $CONFIG_FOLDER/config.toml '.p2p.seeds' ''

    # Disable fast_sync
    dasel put bool -f $CONFIG_FOLDER/config.toml '.fast_sync' 'false'

    # Expose the rpc
    dasel put string -f $CONFIG_FOLDER/config.toml '.rpc.laddr' "tcp://0.0.0.0:26657"

    # minimum-gas-prices config in app.toml, empty string by default
    dasel put string -f $CONFIG_FOLDER/app.toml 'minimum-gas-prices' "0ujuno"
}

if [[ ! -d $CONFIG_FOLDER ]]
then

    install_prerequisites

    echo "Chain ID: $CHAIN_ID"
    echo "Moniker:  $MONIKER"

    echo $MNEMONIC | junod init $MONIKER -o --chain-id=$CHAIN_ID --home $JUNO_HOME
    echo $MNEMONIC | junod keys add my-key --recover --keyring-backend test 2> /dev/null

    ACCOUNT_PUBKEY=$(junod keys show --keyring-backend test my-key --pubkey | dasel -r json '.key' --plain)
    ACCOUNT_ADDRESS=$(junod keys show -a --keyring-backend test my-key --bech acc)
    ACCOUNT_ADDRESS_JSON=$(junod keys show --keyring-backend test my-key --output json | dasel -r json '.pubkey' --plain)
    echo "Account pubkey:  $ACCOUNT_PUBKEY"
    echo "Account address: $ACCOUNT_ADDRESS"

    ACCOUNT_HEX_ADDRESS=$(junod debug pubkey $ACCOUNT_ADDRESS_JSON --home $JUNO_HOME | grep Address | cut -d " " -f 2)    
    ACCOUNT_OPERATOR_ADDRESS=$(junod debug addr $VALIDATOR_HEX_ADDRESS --home $JUNO_HOME | grep Val | cut -d " " -f 3)    

    VALIDATOR_PUBKEY_JSON=$(junod tendermint show-validator --home $JUNO_HOME)
    VALIDATOR_PUBKEY=$(echo $VALIDATOR_PUBKEY_JSON | dasel -r json '.key' --plain)
    VALIDATOR_HEX_ADDRESS=$(junod debug pubkey $VALIDATOR_PUBKEY_JSON --home $JUNO_HOME | grep Address | cut -d " " -f 2)    
    VALIDATOR_ACCOUNT_ADDRESS=$(junod debug addr $VALIDATOR_HEX_ADDRESS --home $JUNO_HOME | grep Acc | cut -d " " -f 3)
    VALIDATOR_OPERATOR_ADDRESS=$(junod debug addr $VALIDATOR_HEX_ADDRESS --home $JUNO_HOME | grep Val | cut -d " " -f 3)    
    # VALIDATOR_CONSENSUS_ADDRESS=$(junod debug bech32-convert $VALIDATOR_OPERATOR_ADDRESS -p junovalcons  --home $JUNO_HOME)   
    VALIDATOR_CONSENSUS_ADDRESS=$(junod tendermint show-address --home $JUNO_HOME)   
    echo "Validator pubkey:  $VALIDATOR_PUBKEY"
    echo "Validator address: $VALIDATOR_ACCOUNT_ADDRESS"
    echo "Validator operator address: $VALIDATOR_OPERATOR_ADDRESS"
    echo "Validator consensus address: $VALIDATOR_CONSENSUS_ADDRESS"    
     
    python3 -u /juno/testnetify.py \
        -i /juno/state_export.json \
        --output $CONFIG_FOLDER/genesis.json \
        -c $CHAIN_ID \
        --validator-hex-address $ACCOUNT_HEX_ADDRESS \
        --validator-operator-address $ACCOUNT_OPERATOR_ADDRESS \
        --validator-consensus-address $VALIDATOR_CONSENSUS_ADDRESS \
        --validator-pubkey $ACCOUNT_PUBKEY \
        --account-pubkey $ACCOUNT_PUBKEY \
        --account-address $ACCOUNT_ADDRESS \
        --prune-ibc

    edit_config
fi

junod validate-genesis --home $JUNO_HOME
junod start --x-crisis-skip-assert-invariants --home $JUNO_HOME