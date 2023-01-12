# first `sh ./scripts/test_node.sh clean`
# then in a new tab:
# sh ./scripts/oracle/run_local_oracle.sh

ORACLE_FILENAME="test_oracle"

cd price-feeder
make install

price-feeder version

cp config.example.toml $HOME/.juno/$ORACLE_FILENAME.toml

# replace gas_price
sed -i 's/0.0001stake/0.025ujuno/g' ~/.juno/$ORACLE_FILENAME.toml

# replace feeder address
sed -i 's/address = "juno1w20tfhnehc33rgtm9tg8gdtea0svn7twfnyqee"/address = "juno1hj5fveer5cjtn4wd6wstzugjfdxzl0xps73ftl"/g' ~/.juno/$ORACLE_FILENAME.toml

# change chain_id
sed -i 's/chain_id = "test-1"/chain_id = "local-1"/g' ~/.juno/$ORACLE_FILENAME.toml
sed -i 's/"chain_id", "test-1"/"chain_id", "local-1"/g' ~/.juno/$ORACLE_FILENAME.toml

VALOPER_ADDR=$(junod q staking validators --node http://localhost:26657 --output json | jq -r '.validators[0].operator_address')

# change validator
sed -i "s/validator = \"junovaloper1w20tfhnehc33rgtm9tg8gdtea0svn7twkwj0zq\"/validator = \"$VALOPER_ADDR\"/g" ~/.juno/$ORACLE_FILENAME.toml

# change 'websocket
sed -i 's/websocket = "stream.binance.com:9443"/websocket = "fstream.binance.com:443"/g' ~/.juno/$ORACLE_FILENAME.toml

# start it 
price-feeder $HOME/.juno/$ORACLE_FILENAME.toml