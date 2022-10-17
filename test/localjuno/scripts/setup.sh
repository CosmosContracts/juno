#!/bin/sh

CHAIN_ID=localjuno
JUNO_HOME=$HOME/.juno
CONFIG_FOLDER=$JUNO_HOME/config
MONIKER=val
STATE='false'

# val - juno1jxa3ksucx7ter57xyuczvmk6qkeqmqvj37g237
MNEMONIC="blame tube add leopard fire next exercise evoke young team payment senior know estate mandate negative actual aware slab drive celery elevator burden utility"

while getopts s flag
do
    case "${flag}" in
        s) STATE='true';;
    esac
done

install_prerequisites () {
    apk add dasel
}

edit_genesis () {

    GENESIS=$CONFIG_FOLDER/genesis.json

    # Update staking module
    dasel put string -f $GENESIS '.app_state.staking.params.bond_denom' 'ujuno'
    dasel put string -f $GENESIS '.app_state.staking.params.unbonding_time' '240s'

    # Update crisis module
    dasel put string -f $GENESIS '.app_state.crisis.constant_fee.denom' 'ujuno'

    # Udpate gov module
    dasel put string -f $GENESIS '.app_state.gov.voting_params.voting_period' '60s'
    dasel put string -f $GENESIS '.app_state.gov.deposit_params.min_deposit.[0].denom' 'ujuno'

    # Update wasm permission (Nobody or Everybody)
    dasel put string -f $GENESIS '.app_state.wasm.params.code_upload_access.permission' "Everybody"
}

add_genesis_accounts () {

    junod add-genesis-account juno1jxa3ksucx7ter57xyuczvmk6qkeqmqvj37g237 100000000000ujuno --home $JUNO_HOME # val
    junod add-genesis-account juno1cyyzpxplxdzkeea7kwsydadg87357qnaf5xk87 100000000000ujuno --home $JUNO_HOME # lo-test1
    junod add-genesis-account juno18s5lynnmx37hq4wlrw9gdn68sg2uxp5rkl63az 100000000000ujuno --home $JUNO_HOME
    junod add-genesis-account juno1qwexv7c6sm95lwhzn9027vyu2ccneaqanu7v8n 100000000000ujuno --home $JUNO_HOME
    junod add-genesis-account juno14hcxlnwlqtq75ttaxf674vk6mafspg8xsprc9l 100000000000ujuno --home $JUNO_HOME
    junod add-genesis-account juno12rr534cer5c0vj53eq4y32lcwguyy7nnnzlhm9 100000000000ujuno --home $JUNO_HOME
    junod add-genesis-account juno1nt33cjd5auzh36syym6azgc8tve0jlvkp6s4rw 100000000000ujuno --home $JUNO_HOME
    junod add-genesis-account juno10qfrpash5g2vk3hppvu45x0g860czur8hqy0hp 100000000000ujuno --home $JUNO_HOME
    junod add-genesis-account juno1f4tvsdukfwh6s9swrc24gkuz23tp8pd38vnlcn 100000000000ujuno --home $JUNO_HOME
    junod add-genesis-account juno1myv43sqgnj5sm4zl98ftl45af9cfzk7nfmke3e 100000000000ujuno --home $JUNO_HOME
    junod add-genesis-account juno14gs9zqh8m49yy9kscjqu9h72exyf295ahp2aec 100000000000ujuno --home $JUNO_HOME # lo-test10

    echo $MNEMONIC | junod keys add $MONIKER --recover --keyring-backend=test --home $JUNO_HOME
    junod gentx $MONIKER 500000000ujuno --keyring-backend=test --chain-id=$CHAIN_ID --home $JUNO_HOME

    junod collect-gentxs --home $JUNO_HOME
}

edit_config () {
    # Remove seeds
    dasel put string -f $CONFIG_FOLDER/config.toml '.p2p.seeds' ''

    # Expose the rpc
    dasel put string -f $CONFIG_FOLDER/config.toml '.rpc.laddr' "tcp://0.0.0.0:26657"
}



if [[ ! -d $CONFIG_FOLDER ]]
then
    echo $MNEMONIC | junod init -o --chain-id=$CHAIN_ID --home $JUNO_HOME --recover $MONIKER
    install_prerequisites
    edit_genesis
    add_genesis_accounts
    edit_config
fi

junod start --home $JUNO_HOME &
# killall junod

wait
