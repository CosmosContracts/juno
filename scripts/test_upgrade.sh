#!/bin/bash
# Run this script to quickly install, setup, and run the current version of juno without docker.
#
# MULTIPLE VALIDATOR SETUP
# BINARY=junodv11 NEW_BINARY=junod TIMEOUT_COMMIT="5000ms" CLEAN=true sh ./scripts/test_upgrade.sh
#
# VALOPER_ADDR=junovaloper1efd63aw40lxf3n4mhf7dzhjkr453axurnh5ze0 FEEDER_ADDRESS=juno1efd63aw40lxf3n4mhf7dzhjkr453axurv2zdzk JUNO_DIR="~/.juno1/" sh ./scripts/oracle/run_local_oracle.sh
# VALOPER_ADDR=junovaloper1hj5fveer5cjtn4wd6wstzugjfdxzl0xp0r8xsx FEEDER_ADDRESS=juno1hj5fveer5cjtn4wd6wstzugjfdxzl0xps73ftl LISTEN_ADDR="0.0.0.0:7172" JUNO_DIR="~/.juno2/" sh ./scripts/oracle/run_local_oracle.sh

export KEY="juno1"
export CHAIN_ID=${CHAIN_ID:-"local-1"}
export KEYALGO="secp256k1"
export KEYRING=${KEYRING:-"test"}

export HOME_DIR=$(eval echo "${HOME_DIR:-"~/.juno1/"}")
export HOME_DIRB=$(eval echo "${HOME_DIRB:-"~/.juno2/"}")
export HOME_DIRC=$(eval echo "${HOME_DIRC:-"~/.juno3/"}")

export BINARY=${BINARY:-junodv11}
export NEW_BINARY=${NEW_BINARY:-junod}
export CLEAN=${CLEAN:-"false"}

export TIMEOUT_COMMIT=${TIMEOUT_COMMIT:-"5s"}

# Validator 1
export RPC=${RPC:-"26657"}
export REST=${REST:-"1317"}
export PROFF=${PROFF:-"6060"}
export P2P=${P2P:-"26656"}
export GRPC=${GRPC:-"9090"}
export GRPC_WEB=${GRPC_WEB:-"9091"}
export ROSETTA=${ROSETTA:-"8080"}

# Validator 2
export RPCB=${RPCB:-"36657"}
export RESTB=${RESTB:-"2317"}
export PROFFB=${PROFFB:-"6061"}
export P2PB=${P2PB:-"36656"}
export GRPCB=${GRPCB:-"8090"}
export GRPC_WEBB=${GRPC_WEBB:-"8091"}
export ROSETTAB=${ROSETTAB:-"8081"}

# Validator 3
export RPCC=${RPCC:-"46657"}
export RESTC=${RESTC:-"3317"}
export PROFFC=${PROFFC:-"7061"}
export P2PC=${P2PC:-"46656"}
export GRPCC=${GRPCC:-"8097"}
export GRPC_WEBC=${GRPC_WEBC:-"8098"}
export ROSETTAC=${ROSETTAC:-"8089"}

# Kill all instances before starting back up
killall $BINARY
killall $NEW_BINARY

# Debugging
# echo "CHAIN_ID=$CHAIN_ID, HOME_DIR=$HOME_DIR, CLEAN=$CLEAN, RPC=$RPC, REST=$REST, PROFF=$PROFF, P2P=$P2P, GRPC=$GRPC, GRPC_WEB=$GRPC_WEB, ROSETTA=$ROSETTA, TIMEOUT_COMMIT=$TIMEOUT_COMMIT"
# echo "CHAIN_ID=$CHAIN_ID, HOME_DIR=$HOME_DIRB, CLEAN=$CLEAN, RPC=$RPC2, REST=$REST2, PROFF=$PROFF2, P2P=$P2P2, GRPC=$GRPC2, GRPC_WEB=$GRPC_WEB2, ROSETTA=$ROSETTA2, TIMEOUT_COMMIT=$TIMEOUT_COMMIT"

command -v junod > /dev/null 2>&1 || { echo >&2 "junod command not found. Ensure this is setup / properly installed in your GOPATH."; exit 1; }
command -v jq > /dev/null 2>&1 || { echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"; exit 1; }

add_keys () {
  NAME=$1
  WORDS=$2
  # add to all home dirs
  echo "$WORDS" | $BINARY --home=$HOME_DIR keys add $NAME --keyring-backend $KEYRING --algo $KEYALGO --recover
  echo "$WORDS" | $BINARY --home=$HOME_DIRB keys add $NAME --keyring-backend $KEYRING --algo $KEYALGO --recover
  echo "$WORDS" | $BINARY --home=$HOME_DIRC keys add $NAME --keyring-backend $KEYRING --algo $KEYALGO --recover
}

add_balances () {
  ADDRESS=$1

  $BINARY --home=$HOME_DIR add-genesis-account $ADDRESS 10000000ujuno,1000utest --keyring-backend $KEYRING
  $BINARY --home=$HOME_DIRB add-genesis-account $ADDRESS 10000000ujuno,1000utest --keyring-backend $KEYRING
  $BINARY --home=$HOME_DIRC add-genesis-account $ADDRESS 10000000ujuno,1000utest --keyring-backend $KEYRING
}

from_scratch () {
  # Fresh install on current branch
  make install

  # remove existing daemon.
  rm -rf $HOME_DIR && echo "Removed $HOME_DIR"  
  rm -rf $HOME_DIRB && echo "Removed $HOME_DIRB"  
  rm -rf $HOME_DIRC && echo "Removed $HOME_DIRC"    

  $BINARY --home=$HOME_DIR init "ahh" --chain-id $CHAIN_ID
  $BINARY --home=$HOME_DIRB init "bee" --chain-id $CHAIN_ID
  $BINARY --home=$HOME_DIRC init "cee" --chain-id $CHAIN_ID

  # juno1efd63aw40lxf3n4mhf7dzhjkr453axurv2zdzk  
  add_keys $KEY "decorate bright ozone fork gallery riot bus exhaust worth way bone indoor calm squirrel merry zero scheme cotton until shop any excess stage laundry"

  # juno1hj5fveer5cjtn4wd6wstzugjfdxzl0xps73ftl  
  add_keys "feeacc" "wealth flavor believe regret funny network recall kiss grape useless pepper cram hint member few certain unveil rather brick bargain curious require crowd raise"

  # juno1g20vre3x9l35rwterkrfw47kyhgypzm5ezewjd
  add_keys "val3" "stable echo above noise tooth master dilemma defense water boost mirror witness quick emotion napkin crowd purity cabbage survey stomach story bounce cake become"  

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
  add_balances juno1efd63aw40lxf3n4mhf7dzhjkr453axurv2zdzk
  add_balances juno1hj5fveer5cjtn4wd6wstzugjfdxzl0xps73ftl
  add_balances juno1g20vre3x9l35rwterkrfw47kyhgypzm5ezewjd

    

  # Gentxs
  GENTX_DEFAULT="1000000ujuno --keyring-backend $KEYRING --chain-id $CHAIN_ID"
  $BINARY gentx $KEY $GENTX_DEFAULT --home $HOME_DIR
  BINARY_1_PEER=$($BINARY tendermint show-node-id --home $HOME_DIR)

  # 2.
  $BINARY gentx feeacc $GENTX_DEFAULT --home $HOME_DIRB
  BINARY_2_PEER=$($BINARY tendermint show-node-id --home $HOME_DIRB)
  cp $HOME_DIRB/config/gentx/*.json $HOME_DIR/config/gentx/gentx-other.json  

  # 3.
  # $BINARY gentx val3 $GENTX_DEFAULT --ip 127.0.0.1 --home $HOME_DIRC 
  # BINARY_3_PEER=$($BINARY tendermint show-node-id --home $HOME_DIRC)
  # cp $HOME_DIRC/config/gentx/*.json $HOME_DIR/config/gentx/gentx-other_2.json  

  # Collect genesis tx
  $BINARY --home=$HOME_DIR collect-gentxs --home $HOME_DIR

  # Run this to ensure junorything worked and that the genesis file is setup correctly
  $BINARY --home=$HOME_DIR validate-genesis

  # copy it from the first node to the second & 3rd  
  cp $HOME_DIR/config/genesis.json $HOME_DIRB/config/genesis.json
  cp $HOME_DIR/config/genesis.json $HOME_DIRC/config/genesis.json
}

# check if CLEAN is not set to false
if [ "$CLEAN" != "false" ]; then
  echo "Starting from a clean state"
  from_scratch
fi


update_config () {
  JUNO_DIR=$1
  tmpRPC=$2
  tmpREST=$3
  tmpPROFF_LADDER=$4
  tmpP2P=$5
  tmpGRPC=$6
  tmpGRPC_WEB=$7
  tmpROSETTA=$8

  CONFIG=$JUNO_DIR/config/config.toml
  APP=$JUNO_DIR/config/app.toml  

  # Opens the RPC endpoint to outside connections
  sed -i 's/laddr = "tcp:\/\/127.0.0.1:26657"/c\laddr = "tcp:\/\/0.0.0.0:'$tmpRPC'"/g' $CONFIG
  sed -i 's/cors_allowed_origins = \[\]/cors_allowed_origins = \["\*"\]/g' $CONFIG

  # REST endpoint
  sed -i 's/address = "tcp:\/\/0.0.0.0:1317"/address = "tcp:\/\/0.0.0.0:'$tmpREST'"/g' $APP
  sed -i 's/enable = false/enable = true/g' $APP

  # replace pprof_laddr = "localhost:6060" binding
  sed -i 's/pprof_laddr = "localhost:6060"/pprof_laddr = "localhost:'$tmpPROFF_LADDER'"/g' $CONFIG

  # change p2p addr laddr = "tcp://0.0.0.0:26656"
  sed -i 's/laddr = "tcp:\/\/0.0.0.0:26656"/laddr = "tcp:\/\/0.0.0.0:'$tmpP2P'"/g' $CONFIG

  # GRPC
  sed -i 's/address = "0.0.0.0:9090"/address = "0.0.0.0:'$tmpGRPC'"/g' $APP
  sed -i 's/address = "0.0.0.0:9091"/address = "0.0.0.0:'$tmpGRPC_WEB'"/g' $APP

  # Rosetta Api
  sed -i 's/address = ":8080"/address = "0.0.0.0:'$tmpROSETTA'"/g' $APP  

  # faster blocks
  sed -i 's/timeout_commit = "5s"/timeout_commit = "'$TIMEOUT_COMMIT'"/g' $CONFIG
}

# update all configs to their respective ports
update_config "$HOME_DIR" "$RPC" "$REST" "$PROFF_LADDER" "$P2P" "$GRPC" "$GRPC_WEB" "$ROSETTA"
update_config "$HOME_DIRB" "$RPCB" "$RESTB" "$PROFF_LADDERB" "$P2PB" "$GRPCB" "$GRPC_WEBB" "$ROSETTAB"
update_config "$HOME_DIRC" "$RPCC" "$RESTC" "$PROFF_LADDERC" "$P2PC" "$GRPCC" "$GRPC_WEBC" "$ROSETTAC"

# Start Nodes
DEFAULT_START_FLAGS="--grpc-web.enable=false --pruning=default --minimum-gas-prices=0ujuno"

PEER_A="$BINARY_1_PEER"@127.0.0.1:$P2P
PEER_B="$BINARY_2_PEER"@127.0.0.1:$P2PB
# PEER_C="$BINARY_3_PEER"@127.0.0.1:$P2PC

$BINARY start --home $HOME_DIR $DEFAULT_START_FLAGS --rpc.laddr="tcp://0.0.0.0:$RPC" --p2p.persistent_peers "$PEER_B" &
$BINARY start --home $HOME_DIRB $DEFAULT_START_FLAGS --rpc.laddr="tcp://0.0.0.0:$RPCB" --p2p.persistent_peers "$PEER_A" --p2p.laddr="tcp://0.0.0.0:$P2PB" --grpc.address="0.0.0.0:$GRPCB" &
# $BINARY start --home $HOME_DIRC $DEFAULT_START_FLAGS --rpc.laddr="tcp://0.0.0.0:$RPCC" --p2p.persistent_peers "$PEER_A" --p2p.laddr="tcp://0.0.0.0:$P2PC" --grpc.address="0.0.0.0:$GRPCC" &

sleep 15

echo -e "\n\n\n\nSUBMIT PROPOSAL"
$BINARY tx gov submit-proposal software-upgrade v12 --title "v12 upgrade test" --description "test upgrade" --deposit 1000000ujuno --upgrade-height 7 --from $KEY --keyring-backend test --home $HOME_DIR --chain-id $CHAIN_ID --yes --broadcast-mode block
echo -e "\n\n\nVOTE"
ID="1" && $BINARY tx gov vote $ID yes --from $KEY --keyring-backend $KEYRING --chain-id $CHAIN_ID --broadcast-mode block --home $HOME_DIR --yes
$BINARY q gov proposal $ID
sleep 30

# better way?
echo -e "\n\n\nKILL ALL JUNODv11"
killall -KILL junodv11
sleep 10

echo -e "\n\n\nSTART NEW"

# start the nodes again as NEW_BINARY
DEFAULT_START_FLAGS="--grpc-web.enable=false --pruning=default --minimum-gas-prices=0ujuno"
$NEW_BINARY start --home $HOME_DIR $DEFAULT_START_FLAGS --rpc.laddr="tcp://0.0.0.0:$RPC" --p2p.persistent_peers "$PEER_B" &
$NEW_BINARY start --home $HOME_DIRB $DEFAULT_START_FLAGS --rpc.laddr="tcp://0.0.0.0:$RPCB" --p2p.persistent_peers "$PEER_A" --p2p.laddr="tcp://0.0.0.0:$P2PB" --grpc.address="0.0.0.0:$GRPCB" &