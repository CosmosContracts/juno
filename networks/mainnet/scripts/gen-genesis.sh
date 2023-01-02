#!/bin/sh
JUNOD_HOME="/tmp/junod$(date +%s)"
CHAIN_ID=juno-1

set -e

echo "...........Init Juno.............."

git clone https://github.com/CosmosContracts/Juno
cd Juno
make build
chmod +x ./build/junod

./build/junod init --chain-id $CHAIN_ID validator --home $JUNOD_HOME

echo "..........Fetching genesis......."
rm -rf $JUNOD_HOME/config/genesis.json
cp ../$CHAIN_ID/genesis-prelaunch.json $JUNOD_HOME/config/genesis.json

echo "..........Collecting gentxs......."
./build/junod collect-gentxs --home $JUNOD_HOME --gentx-dir ../$CHAIN_ID/gentxs

./build/junod validate-genesis --home $JUNOD_HOME

cp $JUNOD_HOME/config/genesis.json ../$CHAIN_ID/genesis.json
jq -S -c -M '' ../$CHAIN_ID/genesis.json | shasum -a 256 > ../$CHAIN_ID/checksum.txt

echo "..........Starting node......."
./build/junod start --home $JUNOD_HOME &

sleep 5s

echo "...Cleaning the stuff..."
killall junod >/dev/null 2>&1
rm -rf $JUNOD_HOME >/dev/null 2>&1

cd ..
rm -rf Juno
