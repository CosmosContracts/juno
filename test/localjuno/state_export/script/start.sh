#!/bin/sh
set -e 
set -o pipefail

JUNO_HOME=$HOME/.juno
CONFIG_FOLDER=$JUNO_HOME/config

# val - juno1jxa3ksucx7ter57xyuczvmk6qkeqmqvj37g237
DEFAULT_MNEMONIC="blame tube add leopard fire next exercise evoke young team payment senior know estate mandate negative actual aware slab drive celery elevator burden utility"
DEFAULT_CHAIN_ID="localjuno"
DEFAULT_MONIKER="val"

# Override default values with environment variables
MNEMONIC=${MNEMONIC:-$DEFAULT_MNEMONIC}
CHAIN_ID=${CHAIN_ID:-$DEFAULT_CHAIN_ID}
MONIKER=${MONIKER:-$DEFAULT_MONIKER}

# Debug print in the docker container, then exit
# echo $(ls /juno); exit 1

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
}

# if [[ ! -d $CONFIG_FOLDER ]]
# then

    install_prerequisites

    echo "Chain ID: $CHAIN_ID"
    echo "Moniker:  $MONIKER"

    echo $MNEMONIC | junod init $MONIKER -o --chain-id=$CHAIN_ID --home $JUNO_HOME &2> /dev/null
    echo $MNEMONIC | junod keys add my-key --recover --keyring-backend test &2> /dev/null   

    ACCOUNT_PUBKEY=$(junod keys show --keyring-backend test my-key --pubkey | dasel -r json '.key' --plain)
    ACCOUNT_ADDRESS=$(junod keys show -a --keyring-backend test my-key --bech acc)
    echo "Account pubkey:  $ACCOUNT_PUBKEY"
    echo "Account address: $ACCOUNT_ADDRESS"
    
    # create a validator    
    junod add-genesis-account $ACCOUNT_ADDRESS 100000000000ujuno --home $JUNO_HOME # val
    # junod gentx my-key --moniker=$MONIKER 500000000ujuno --keyring-backend=test --chain-id=$CHAIN_ID --home $JUNO_HOME
    junod collect-gentxs --home $JUNO_HOME     

    # echo $(ls /juno); exit 1    

    # TODO:  $JUNO_HOME/config/priv_validator_key.json not found here, but we can't run this without CONFIG_FOLDER being new? 
    # So this always fails. Are we meant to call the ../../scripts/setup.sh first within here?
    VALIDATOR_PUBKEY_JSON=$(junod tendermint show-validator --home $JUNO_HOME)
    VALIDATOR_PUBKEY=$(echo $VALIDATOR_PUBKEY_JSON | dasel -r json '.key' --plain)
    VALIDATOR_HEX_ADDRESS=$(junod debug pubkey $VALIDATOR_PUBKEY_JSON --home $JUNO_HOME | grep Address | cut -d " " -f 2)    
    VALIDATOR_ACCOUNT_ADDRESS=$(junod debug addr $VALIDATOR_HEX_ADDRESS --home $JUNO_HOME | grep Acc | cut -d " " -f 3)
    VALIDATOR_OPERATOR_ADDRESS=$(junod debug addr $VALIDATOR_HEX_ADDRESS --home $JUNO_HOME | grep Val | cut -d " " -f 3)    
    VALIDATOR_CONSENSUS_ADDRESS=$(junod debug bech32-convert $VALIDATOR_OPERATOR_ADDRESS -p junovalcons  --home $JUNO_HOME)   

    

    echo "Validator pubkey:  $VALIDATOR_PUBKEY"
    echo "Validator address: $VALIDATOR_ACCOUNT_ADDRESS"
    echo "Validator operator address: $VALIDATOR_OPERATOR_ADDRESS"
    echo "Validator consensus address: $VALIDATOR_CONSENSUS_ADDRESS"    

    python3 -u /juno/testnetify.py \
    -i /juno/state_export.json \
    -o $CONFIG_FOLDER/genesis.json \
    -c $CHAIN_ID \
    --validator-hex-address $VALIDATOR_HEX_ADDRESS \
    --validator-operator-address $VALIDATOR_OPERATOR_ADDRESS \
    --validator-consensus-address $VALIDATOR_CONSENSUS_ADDRESS \
    --validator-pubkey $VALIDATOR_PUBKEY \
    --account-pubkey $ACCOUNT_PUBKEY \
    --account-address $ACCOUNT_ADDRESS \
    --prune-ibc

    edit_config
# fi

junod start --home $JUNO_HOME --x-crisis-skip-assert-invariants
