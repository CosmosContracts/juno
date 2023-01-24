# This script is boilerplate to easily spin up a remote, multi node testnet in <5 minutes as 1 person.
# In the future this could be entirely autoamted with screen / ssh commands, but for now it's good enough
#
# apt update && apt -y upgrade && apt -y install make gcc screen htop git snapd
# sudo snap install go --classic
# # or another version like so:
# curl -OL https://golang.org/dl/go1.18.linux-amd64.tar.gz
# sudo tar -C /usr/local -xvf go1.18.linux-amd64.tar.gz
# rm /usr/bin/go
# bash
# 
# # Install JUNOv11 and v12
# git clone https://github.com/cosmoscontracts/juno.git && cd juno
# git checkout v11.0.3 && make install && mv ~/go/bin/junod ~/go/bin/junodv11 && chmod +x ~/go/bin/junodv11
# git checkout v12.0.0-alpha && make install && chmod +x ~/go/bin/junod
#
# ssh-copy-id -i ~/.ssh/id_ed25519.pub root@5.161.80.115
# ssh-copy-id -i ~/.ssh/id_ed25519.pub root@162.55.180.146
# ssh-copy-id -i ~/.ssh/id_ed25519.pub root@95.216.144.17
# rm -rf ~/.juno1/
#
# Run manually depending on where it says local or not. Copy from here all the way down to ====/ALL==== on every machine

export KEY="juno1"
export CHAIN_ID=${CHAIN_ID:-"local-1"}
export KEYALGO="secp256k1"
export KEYRING=${KEYRING:-"test"}
export HOME_DIR=$(eval echo "${HOME_DIR:-"~/.juno1/"}")
export RPC=${RPC:-"26657"}
export REST=${REST:-"1317"}
export PROFF=${PROFF:-"6060"}
export P2P=${P2P:-"26656"}
export GRPC=${GRPC:-"9090"}
export GRPC_WEB=${GRPC_WEB:-"9091"}
export ROSETTA=${ROSETTA:-"8080"}
export TIMEOUT_COMMIT=5s
BINARY=${BINARY:-"junodv11"}
NEW_BINARY="junod"
add_keys () {
  NAME=$1
  WORDS=$2
  echo "$WORDS" | $BINARY --home=$HOME_DIR keys add $NAME --keyring-backend $KEYRING --algo $KEYALGO --recover  
}
add_balances () {
  ADDRESS=$1
  $BINARY --home=$HOME_DIR add-genesis-account $ADDRESS 10000000ujuno,1000utest --keyring-backend $KEYRING
}
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
update_test_genesis () {
  # update_test_genesis '.consensus_params["block"]["max_gas"]="100000000"'
  cat $HOME_DIR/config/genesis.json | jq "$1" > $HOME_DIR/config/tmp_genesis.json && mv $HOME_DIR/config/tmp_genesis.json $HOME_DIR/config/genesis.json
}

rm -rf $HOME_DIR && echo "Removed $HOME_DIR"  
$BINARY init "$HOSTNAME" --chain-id $CHAIN_ID --home=$HOME_DIR

# juno1efd63aw40lxf3n4mhf7dzhjkr453axurv2zdzk  
add_keys juno1 "decorate bright ozone fork gallery riot bus exhaust worth way bone indoor calm squirrel merry zero scheme cotton until shop any excess stage laundry"
# juno1hj5fveer5cjtn4wd6wstzugjfdxzl0xps73ftl  
add_keys "feeacc" "wealth flavor believe regret funny network recall kiss grape useless pepper cram hint member few certain unveil rather brick bargain curious require crowd raise"
# juno1g20vre3x9l35rwterkrfw47kyhgypzm5ezewjd
add_keys "val3" "stable echo above noise tooth master dilemma defense water boost mirror witness quick emotion napkin crowd purity cabbage survey stomach story bounce cake become" 
add_balances juno1efd63aw40lxf3n4mhf7dzhjkr453axurv2zdzk
add_balances juno1hj5fveer5cjtn4wd6wstzugjfdxzl0xps73ftl
add_balances juno1g20vre3x9l35rwterkrfw47kyhgypzm5ezewjd
export PATH=$PATH:~/go/bin
update_config "$HOME_DIR" "$RPC" "$REST" "$PROFF_LADDER" "$P2P" "$GRPC" "$GRPC_WEB" "$ROSETTA"
# ====/ALL====


# LOCAL (generates the genesis file params)
$BINARY init "local" --chain-id $CHAIN_ID --home=$HOME_DIR
update_test_genesis '.consensus_params["block"]["max_gas"]="100000000"'
update_test_genesis '.app_state["gov"]["voting_params"]["voting_period"]="15s"'
update_test_genesis '.app_state["staking"]["params"]["bond_denom"]="ujuno"'  
update_test_genesis '.app_state["bank"]["params"]["send_enabled"]=[{"denom": "ujuno","enabled": true}]'
update_test_genesis '.app_state["mint"]["params"]["mint_denom"]="ujuno"'  
update_test_genesis '.app_state["gov"]["deposit_params"]["min_deposit"]=[{"denom": "ujuno","amount": "1000000"}]'
update_test_genesis '.app_state["crisis"]["constant_fee"]={"denom": "ujuno","amount": "1000"}'  
update_test_genesis '.app_state["tokenfactory"]["params"]["denom_creation_fee"]=[{"denom":"ujuno","amount":"100"}]'
update_test_genesis '.app_state["feeshare"]["params"]["allowed_denoms"]=["ujuno"]'
$BINARY add-ica-config --home=$HOME_DIR




# Gentxs, on each

# first - juno-1
NODE_IP=`curl ipinfo.io/ip`
GENTX_DEFAULT="1000000ujuno --keyring-backend $KEYRING --chain-id $CHAIN_ID"
$BINARY gentx juno1 $GENTX_DEFAULT --home $HOME_DIR --ip $NODE_IP

# second - juno-2-VDS
NODE_IP=`curl ipinfo.io/ip`
GENTX_DEFAULT="1000000ujuno --keyring-backend $KEYRING --chain-id $CHAIN_ID"
$BINARY gentx feeacc $GENTX_DEFAULT --home $HOME_DIR --ip $NODE_IP

# third - finland
NODE_IP=`curl ipinfo.io/ip`
GENTX_DEFAULT="1000000ujuno --keyring-backend $KEYRING --chain-id $CHAIN_ID"
$BINARY gentx val3 $GENTX_DEFAULT --home $HOME_DIR --ip $NODE_IP



# LOCAL
THIS_DIR=/home/reece/Desktop/multinode # CHANGE THIS TO THIS REPO
scp root@5.161.80.115:/root/.juno1/config/gentx/*.json $THIS_DIR/node1.json
scp root@162.55.180.146:/root/.juno1/config/gentx/*.json $THIS_DIR/node2.json
scp root@95.216.144.17:/root/.juno1/config/gentx/*.json $THIS_DIR/node3.json

# LOCAL - GET PEERS FROM THE GENTXS
BINARY_1_PEER=$(cat $THIS_DIR/node1.json | jq -r '.body.memo') && echo $BINARY_1_PEER
BINARY_2_PEER=$(cat $THIS_DIR/node2.json | jq -r '.body.memo') && echo $BINARY_2_PEER
BINARY_3_PEER=$(cat $THIS_DIR/node3.json | jq -r '.body.memo') && echo $BINARY_3_PEER

mkdir -p $HOME_DIR/config/gentx
cp $THIS_DIR/node1.json $HOME_DIR/config/gentx/
cp $THIS_DIR/node2.json $HOME_DIR/config/gentx/
cp $THIS_DIR/node3.json $HOME_DIR/config/gentx/
$BINARY collect-gentxs --home $HOME_DIR
$BINARY --home=$HOME_DIR validate-genesis 


scp -r $HOME_DIR/config/genesis.json root@5.161.80.115:/root/.juno1/config/genesis.json
scp -r $HOME_DIR/config/genesis.json root@162.55.180.146:/root/.juno1/config/genesis.json
scp -r $HOME_DIR/config/genesis.json root@95.216.144.17:/root/.juno1/config/genesis.json



#   LOCAL, then copy paste to their respective servers
# juno-1
echo -e "\n\n\n$BINARY" start --home /root/.juno1/ --grpc-web.enable=false --pruning=default --minimum-gas-prices=0ujuno --rpc.laddr="tcp://0.0.0.0:$RPC" --p2p.persistent_peers "$BINARY_2_PEER,$BINARY_3_PEER"
# juno-2-VDS
echo -e "\n\n\n$BINARY" start --home /root/.juno1/ --grpc-web.enable=false --pruning=default --minimum-gas-prices=0ujuno --rpc.laddr="tcp://0.0.0.0:$RPC" --p2p.persistent_peers "$BINARY_1_PEER,$BINARY_3_PEER"
# finland
echo -e "\n\n\n$BINARY" start --home /root/.juno1/ --grpc-web.enable=false --pruning=default --minimum-gas-prices=0ujuno --rpc.laddr="tcp://0.0.0.0:$RPC" --p2p.persistent_peers "$BINARY_2_PEER,$BINARY_1_PEER"




# LOCAL
NODE="--node http://5.161.80.115:26657"
$BINARY tx gov submit-proposal software-upgrade v12 --title "v12 upgrade test" --description "test upgrade" $NODE --deposit 1000000ujuno --upgrade-height 40 --from $KEY --keyring-backend test --home $HOME_DIR --chain-id $CHAIN_ID --yes --broadcast-mode block
echo -e "\n\n\nVOTE"
ID="1" && $BINARY tx gov vote $ID yes --from $KEY --keyring-backend $KEYRING --chain-id $CHAIN_ID --broadcast-mode block --home $HOME_DIR $NODE --yes && $BINARY tx gov vote $ID yes --from val3 --keyring-backend $KEYRING --chain-id $CHAIN_ID --broadcast-mode block --home $HOME_DIR $NODE --yes
$BINARY q gov proposal $ID $NODE

# STOP EACH, then run this

# juno-1
echo -e "\n\n\n$NEW_BINARY" start --home /root/.juno1/ --grpc-web.enable=false --pruning=default --minimum-gas-prices=0ujuno --rpc.laddr="tcp://0.0.0.0:$RPC" --p2p.persistent_peers "$BINARY_2_PEER,$BINARY_3_PEER"
# juno-2-VDS
echo -e "\n\n\n$NEW_BINARY" start --home /root/.juno1/ --grpc-web.enable=false --pruning=default --minimum-gas-prices=0ujuno --rpc.laddr="tcp://0.0.0.0:$RPC" --p2p.persistent_peers "$BINARY_1_PEER,$BINARY_3_PEER"
# finland
echo -e "\n\n\n$NEW_BINARY" start --home /root/.juno1/ --grpc-web.enable=false --pruning=default --minimum-gas-prices=0ujuno --rpc.laddr="tcp://0.0.0.0:$RPC" --p2p.persistent_peers "$BINARY_2_PEER,$BINARY_1_PEER"


# arch
junod tx bank send juno1 juno1efd63aw40lxf3n4mhf7dzhjkr453axurv2zdzk 1000000ujuno --keyring-backend test --chain-id local-1 --home ~/.juno1 --yes --broadcast-mode block $NODE --fees 500ujuno

# store contracts?