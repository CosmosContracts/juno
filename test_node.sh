# Ensure juno is installed first.

# account with 1ujuno of funds at genesis for fee pay testing
# juno1efd63aw40lxf3n4mhf7dzhjkr453axurv2zdzk
# key = feeacc 

KEY="juno1"
CHAINID="juno-t1"
MONIKER="localjuno"
KEYALGO="secp256k1"
KEYRING="test" # export juno_KEYRING="TEST"
LOGLjunoL="info"
TRACE="" # "--trace"

junod config keyring-backend $KEYRING
junod config chain-id $CHAINID
# junod config output "json"

command -v jq > /dev/null 2>&1 || { echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"; exit 1; }

from_scratch () {

  make install

  # remove existing daemon
  rm -rf ~/.juno/* 

  # if $KEY exists it should be deleted  
  # juno1hj5fveer5cjtn4wd6wstzugjfdxzl0xpysfwwn
  echo "decorate bright ozone fork gallery riot bus exhaust worth way bone indoor calm squirrel merry zero scheme cotton until shop any excess stage laundry" | junod keys add $KEY --keyring-backend $KEYRING --algo $KEYALGO --recover
  echo "wealth flavor believe regret funny network recall kiss grape useless pepper cram hint member few certain unveil rather brick bargain curious require crowd raise" | junod keys add feeacc --keyring-backend $KEYRING --algo $KEYALGO --recover
  # Set moniker and chain-id for Craft
  junod init $MONIKER --chain-id $CHAINID 

  # Function updates the config based on a jq argument as a string
  update_test_genesis () {
    # update_test_genesis '.consensus_params["block"]["max_gas"]="100000000"'
    cat $HOME/.juno/config/genesis.json | jq "$1" > $HOME/.juno/config/tmp_genesis.json && mv $HOME/.juno/config/tmp_genesis.json $HOME/.juno/config/genesis.json
  }

  # Set gas limit in genesis
  update_test_genesis '.consensus_params["block"]["max_gas"]="100000000"'
  update_test_genesis '.app_state["gov"]["voting_params"]["voting_period"]="15s"'

  # Change chain options to use EXP as the staking denom for craft
  update_test_genesis '.app_state["staking"]["params"]["bond_denom"]="ujuno"'
  # update_test_genesis '.app_state["bank"]["params"]["send_enabled"]=[{"denom": "ujuno","enabled": false}]'
#   update_test_genesis '.app_state["staking"]["params"]["min_commission_rate"]="0.100000000000000000"'    

  # update from token -> ucraft
  update_test_genesis '.app_state["mint"]["params"]["mint_denom"]="ujuno"'  
  update_test_genesis '.app_state["gov"]["deposit_params"]["min_deposit"]=[{"denom": "ujuno","amount": "1000000"}]' # 1 juno right now
  update_test_genesis '.app_state["crisis"]["constant_fee"]={"denom": "ujuno","amount": "1000"}'  

  update_test_genesis '.app_state["tokenfactory"]["params"]["denom_creation_fee"]=[{"denom":"ujuno","amount":"100"}]'

  # Allocate genesis accounts
  # 10 juno (1 of which is used for validator)
  junod add-genesis-account $KEY 10000000ujuno --keyring-backend $KEYRING
  junod add-genesis-account feeacc 1000000ujuno --keyring-backend $KEYRING

  # create gentx with 1 juno
  junod gentx $KEY 1000000ujuno --keyring-backend $KEYRING --chain-id $CHAINID

  # Collect genesis tx
  junod collect-gentxs

  # Run this to ensure junorything worked and that the genesis file is setup correctly
  junod validate-genesis
}

from_scratch

# Opens the RPC endpoint to outside connections
sed -i '/laddr = "tcp:\/\/127.0.0.1:26657"/c\laddr = "tcp:\/\/0.0.0.0:26657"' ~/.juno/config/config.toml
sed -i 's/cors_allowed_origins = \[\]/cors_allowed_origins = \["\*"\]/g' ~/.juno/config/config.toml
sed -i 's/enable = false/enable = true/g' ~/.juno/config/app.toml
# cors_allowed_origins = []


# # Start the node (remove the --pruning=nothing flag if historical queries are not needed)
junod start --pruning=nothing  --minimum-gas-prices=0ujuno #--mode validator     