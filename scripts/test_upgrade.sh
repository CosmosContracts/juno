#!/bin/bash
# Run this script to quickly install, setup, and run the current version of juno without docker.
#
# MULTIPLE VALIDATOR SETUP
# BINARY=junodv11 NEW_BINARY=junod TIMEOUT_COMMIT="5000ms" CLEAN=true sh ./scripts/test_upgrade.sh

export KEY="juno1"
export CHAIN_ID=${CHAIN_ID:-"local-3"}
export MONIKER="localjuno"
export KEYALGO="secp256k1"
export KEYRING=${KEYRING:-"test"}
export HOME_DIR=$(eval echo "${HOME_DIR:-"~/.juno/"}")
export HOME_DIRB=$(eval echo "${HOME_DIRB:-"~/.juno-2/"}")

export BINARY=${BINARY:-junodv11}
export NEW_BINARY=${NEW_BINARY:-junod}
export CLEAN=${CLEAN:-"false"}

export TIMEOUT_COMMIT=${TIMEOUT_COMMIT:-"5s"}

export RPC=${RPC:-"26657"}
export REST=${REST:-"1317"}
export PROFF=${PROFF:-"6060"}
export P2P=${P2P:-"26656"}
export GRPC=${GRPC:-"9090"}
export GRPC_WEB=${GRPC_WEB:-"9091"}
export ROSETTA=${ROSETTA:-"8080"}

export RPCB=${RPCB:-"36657"}
export RESTB=${RESTB:-"2317"}
export PROFFB=${PROFFB:-"6061"}
export P2PB=${P2PB:-"36656"}
export GRPCB=${GRPCB:-"8090"}
export GRPC_WEBB=${GRPC_WEBB:-"8091"}
export ROSETTAB=${ROSETTAB:-"8081"}

# Kill all instances before starting back up
killall -KILL $BINARY

junod config keyring-backend $KEYRING
junod config chain-id $CHAIN_ID

# Debugging
echo "CHAIN_ID=$CHAIN_ID, HOME_DIR=$HOME_DIR, CLEAN=$CLEAN, RPC=$RPC, REST=$REST, PROFF=$PROFF, P2P=$P2P, GRPC=$GRPC, GRPC_WEB=$GRPC_WEB, ROSETTA=$ROSETTA, TIMEOUT_COMMIT=$TIMEOUT_COMMIT"
echo "CHAIN_ID=$CHAIN_ID, HOME_DIR=$HOME_DIRB, CLEAN=$CLEAN, RPC=$RPC2, REST=$REST2, PROFF=$PROFF2, P2P=$P2P2, GRPC=$GRPC2, GRPC_WEB=$GRPC_WEB2, ROSETTA=$ROSETTA2, TIMEOUT_COMMIT=$TIMEOUT_COMMIT"

command -v junod > /dev/null 2>&1 || { echo >&2 "junod command not found. Ensure this is setup / properly installed in your GOPATH."; exit 1; }
command -v jq > /dev/null 2>&1 || { echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"; exit 1; }

from_scratch () {
  # Fresh install on current branch
  make install

  # remove existing daemon.
  rm -rf $HOME_DIR && echo "Removed $HOME_DIR"  
  rm -rf $HOME_DIRB && echo "Removed $HOME_DIRB"  
  
  # juno1efd63aw40lxf3n4mhf7dzhjkr453axurv2zdzk
  echo "decorate bright ozone fork gallery riot bus exhaust worth way bone indoor calm squirrel merry zero scheme cotton until shop any excess stage laundry" | $BINARY --home=$HOME_DIR keys add $KEY --keyring-backend $KEYRING --algo $KEYALGO --recover
  echo "decorate bright ozone fork gallery riot bus exhaust worth way bone indoor calm squirrel merry zero scheme cotton until shop any excess stage laundry" | $BINARY --home=$HOME_DIRB keys add $KEY --keyring-backend $KEYRING --algo $KEYALGO --recover
  # juno1hj5fveer5cjtn4wd6wstzugjfdxzl0xps73ftl
  echo "wealth flavor believe regret funny network recall kiss grape useless pepper cram hint member few certain unveil rather brick bargain curious require crowd raise" | $BINARY --home=$HOME_DIR keys add feeacc --keyring-backend $KEYRING --algo $KEYALGO --recover
  echo "wealth flavor believe regret funny network recall kiss grape useless pepper cram hint member few certain unveil rather brick bargain curious require crowd raise" | $BINARY --home=$HOME_DIRB keys add feeacc --keyring-backend $KEYRING --algo $KEYALGO --recover
  
  $BINARY --home=$HOME_DIR init $MONIKER --chain-id $CHAIN_ID
  $BINARY --home=$HOME_DIRB init "other" --chain-id $CHAIN_ID

  # Function updates the config based on a jq argument as a string
  update_test_genesis () {
    # update_test_genesis '.consensus_params["block"]["max_gas"]="100000000"'
    cat $HOME_DIR/config/genesis.json | jq "$1" > $HOME_DIR/config/tmp_genesis.json && mv $HOME_DIR/config/tmp_genesis.json $HOME_DIR/config/genesis.json
  }

  # Set gas limit in genesis
  update_test_genesis '.consensus_params["block"]["max_gas"]="100000000"'
  update_test_genesis '.app_state["gov"]["voting_params"]["voting_period"]="15s"'

  update_test_genesis '.app_state["staking"]["params"]["bond_denom"]="ujuno"'  
  update_test_genesis '.app_state["bank"]["params"]["send_enabled"]=[{"denom": "ujuno","enabled": true}]'
  # update_test_genesis '.app_state["staking"]["params"]["min_commission_rate"]="0.100000000000000000"' # sdk 46 only   

  update_test_genesis '.app_state["mint"]["params"]["mint_denom"]="ujuno"'  
  update_test_genesis '.app_state["gov"]["deposit_params"]["min_deposit"]=[{"denom": "ujuno","amount": "1000000"}]'
  update_test_genesis '.app_state["crisis"]["constant_fee"]={"denom": "ujuno","amount": "1000"}'  

  update_test_genesis '.app_state["tokenfactory"]["params"]["denom_creation_fee"]=[{"denom":"ujuno","amount":"100"}]'

  update_test_genesis '.app_state["feeshare"]["params"]["allowed_denoms"]=["ujuno"]'


  # Allocate genesis accounts
  $BINARY --home=$HOME_DIR add-genesis-account $KEY 10000000ujuno,1000utest --keyring-backend $KEYRING
  $BINARY --home=$HOME_DIR add-genesis-account juno1efd63aw40lxf3n4mhf7dzhjkr453axurv2zdzk 10000000ujuno,1000utest --keyring-backend $KEYRING
  $BINARY --home=$HOME_DIRB add-genesis-account feeacc 10000000ujuno,1000utest --keyring-backend $KEYRING  

  # since we copy anyways
  # /home/reece/.juno/config/gentx/gentx-5eb04a48391716aee064ed4d1245b14ddd7997ad.json
  # TODO; allow ip to change via config variable for public access
  $BINARY --home=$HOME_DIR gentx $KEY 1000000ujuno --keyring-backend $KEYRING --chain-id $CHAIN_ID --home $HOME_DIR --ip 127.0.0.1
  BINARY_1_PEER=$($BINARY tendermint show-node-id --home $HOME_DIR)
  # ls $HOME_DIR/config/gentx

  $BINARY gentx feeacc 1000000ujuno --keyring-backend $KEYRING --chain-id $CHAIN_ID --home $HOME_DIRB --ip 127.0.0.1
  BINARY_2_PEER=$($BINARY tendermint show-node-id --home $HOME_DIRB) && echo "$BINARY_2_PEER"
  cp $HOME_DIRB/config/gentx/*.json $HOME_DIR/config/gentx/gentx-other.json  

  # Collect genesis tx
  $BINARY --home=$HOME_DIR collect-gentxs --home $HOME_DIR

  # Run this to ensure junorything worked and that the genesis file is setup correctly
  $BINARY --home=$HOME_DIR validate-genesis

  # copy it from the first node to the second
  cp $HOME_DIR/config/genesis.json $HOME_DIRB/config/genesis.json
}

# check if CLEAN is not set to false
if [ "$CLEAN" != "false" ]; then
  echo "Starting from a clean state"
  from_scratch
fi

# exit 1

# echo "Starting node..."

# FIRST FOLDER
# Opens the RPC endpoint to outside connections
sed -i 's/laddr = "tcp:\/\/127.0.0.1:26657"/c\laddr = "tcp:\/\/0.0.0.0:'$RPC'"/g' $HOME_DIR/config/config.toml
sed -i 's/cors_allowed_origins = \[\]/cors_allowed_origins = \["\*"\]/g' $HOME_DIR/config/config.toml
# REST endpoint
sed -i 's/address = "tcp:\/\/0.0.0.0:1317"/address = "tcp:\/\/0.0.0.0:'$REST'"/g' $HOME_DIR/config/app.toml
sed -i 's/enable = false/enable = true/g' $HOME_DIR/config/app.toml
# replace pprof_laddr = "localhost:6060" binding
sed -i 's/pprof_laddr = "localhost:6060"/pprof_laddr = "localhost:'$PROFF_LADDER'"/g' $HOME_DIR/config/config.toml
# change p2p addr laddr = "tcp://0.0.0.0:26656"
sed -i 's/laddr = "tcp:\/\/0.0.0.0:26656"/laddr = "tcp:\/\/0.0.0.0:'$P2P'"/g' $HOME_DIR/config/config.toml
# GRPC
sed -i 's/address = "0.0.0.0:9090"/address = "0.0.0.0:'$GRPC'"/g' $HOME_DIR/config/app.toml
sed -i 's/address = "0.0.0.0:9091"/address = "0.0.0.0:'$GRPC_WEB'"/g' $HOME_DIR/config/app.toml
# Rosetta Api
sed -i 's/address = ":8080"/address = "0.0.0.0:'$ROSETTA'"/g' $HOME_DIR/config/app.toml
# pruning (fix panic)
sed -i 's/pruning = "default"/pruning = "nothing"/g' $HOME_DIR/config/app.toml
# faster blocks
sed -i 's/timeout_commit = "5s"/timeout_commit = "'$TIMEOUT_COMMIT'"/g' $HOME_DIR/config/config.toml


# SECOND FOLDER
sed -i 's/laddr = "tcp:\/\/127.0.0.1:26657"/c\laddr = "tcp:\/\/0.0.0.0:'$RPCB'"/g' $HOME_DIRB/config/config.toml
sed -i 's/cors_allowed_origins = \[\]/cors_allowed_origins = \["\*"\]/g' $HOME_DIRB/config/config.toml
# REST endpoint
sed -i 's/address = "tcp:\/\/0.0.0.0:1317"/address = "tcp:\/\/0.0.0.0:'$RESTB'"/g' $HOME_DIRB/config/app.toml
sed -i 's/enable = false/enable = true/g' $HOME_DIRB/config/app.toml
# replace pprof_laddr = "localhost:6060" binding
sed -i 's/pprof_laddr = "localhost:6060"/pprof_laddr = "localhost:'$PROFF_LADDERB'"/g' $HOME_DIRB/config/config.toml
# change p2p addr laddr = "tcp://0.0.0.0:26656"
sed -i 's/laddr = "tcp:\/\/0.0.0.0:26656"/laddr = "tcp:\/\/0.0.0.0:'$P2PB'"/g' $HOME_DIRB/config/config.toml
# GRPC
sed -i 's/address = "0.0.0.0:9090"/address = "0.0.0.0:'$GRPCB'"/g' $HOME_DIRB/config/app.toml
sed -i 's/address = "0.0.0.0:9091"/address = "0.0.0.0:'$GRPC_WEBB'"/g' $HOME_DIRB/config/app.toml
# Rosetta Api
sed -i 's/address = ":8080"/address = "0.0.0.0:'$ROSETTAB'"/g' $HOME_DIRB/config/app.toml
# pruning (fix panic)
sed -i 's/pruning = "default"/pruning = "nothing"/g' $HOME_DIRB/config/app.toml

# faster blocks
sed -i 's/timeout_commit = "5s"/timeout_commit = "'$TIMEOUT_COMMIT'"/g' $HOME_DIRB/config/config.toml

# # Start the node with 0 gas fees
$BINARY start --home $HOME_DIR --pruning=nothing  --minimum-gas-prices=0ujuno --rpc.laddr="tcp://0.0.0.0:$RPC" --p2p.persistent_peers "$BINARY_2_PEER"@127.0.0.1:$P2PB &

# start with peer of the other
$BINARY start --home $HOME_DIRB --pruning=nothing  --minimum-gas-prices=0ujuno --rpc.laddr="tcp://0.0.0.0:$RPCB" --p2p.persistent_peers "$BINARY_1_PEER"@127.0.0.1:$P2P --p2p.laddr="tcp://0.0.0.0:$P2PB" --grpc-web.enable=false --grpc.address="0.0.0.0:$GRPCB" &


# killall junod && killall junodv11

sleep 10

# submit prop tio vote and halt eventually
echo -e "\n\n\n\nSUBMIT PROPOSAL"
$BINARY tx gov submit-proposal software-upgrade v12 --title "v12 upgrade test" --description "test upgrade" --deposit 1000000ujuno --upgrade-height 7 --from $KEY --keyring-backend test --home $HOME_DIR --chain-id $CHAIN_ID --yes --broadcast-mode block
echo -e "\n\n\nVOTE"
ID="1" && $BINARY tx gov vote $ID yes --from $KEY --keyring-backend $KEYRING --chain-id $CHAIN_ID --broadcast-mode block --yes
$BINARY q gov proposal $ID
sleep 30

# better way?
echo -e "\n\n\nKILL ALL JUNOD"
killall -9 junod & killall -9 junodv11
sleep 2

echo -e "\n\n\nSTART NEW"
# start both again in the background
$NEW_BINARY start --pruning=nothing  --minimum-gas-prices=0ujuno --rpc.laddr="tcp://0.0.0.0:$RPC" --p2p.persistent_peers="$BINARY_2_PEER"@127.0.0.1:$P2PB &

# start with peer of the other
$NEW_BINARY start --home $HOME_DIRB --pruning=nothing  --minimum-gas-prices=0ujuno --rpc.laddr="tcp://0.0.0.0:$RPCB" --p2p.persistent_peers "$BINARY_1_PEER"@127.0.0.1:$P2P --p2p.laddr="tcp://0.0.0.0:$P2PB" --grpc-web.enable=false --grpc.address="0.0.0.0:$GRPCB" &



doUpgrade() {
  # Run this block manually in your terminal

  # Waitm then we will run these after it halts at height  

  
  # junod q globalfee minimum-gas-prices --node http://localhost:26657

  $NEWBINARY tx gov submit-proposal param-change ~/Desktop/Work/paramChange.json --from $KEY --keyring-backend $KEYRING --chain-id $CHAINID --yes --broadcast-mode block --node http://localhost:26657 --fees 600ujuno
  ID="1"
  $NEWBINARY tx gov vote $ID yes --from $KEY --keyring-backend $KEYRING --chain-id $CHAINID --yes --broadcast-mode block
  $NEWBINARY q gov proposal $ID

  junod tx bank send $KEY juno1efd63aw40lxf3n4mhf7dzhjkr453axurv2zdzk 1ujuno --keyring-backend $KEYRING --chain-id $CHAINID --yes --broadcast-mode block --fees 500ujuno --gas 150000
}