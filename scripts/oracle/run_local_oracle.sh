# first `sh ./scripts/test_node.sh clean`
# then in a new tab:
# FEEDER_ADDRESS=juno1hj5fveer5cjtn4wd6wstzugjfdxzl0xps73ftl JUNO_DIR="~/.juno1/" sh ./scripts/oracle/run_local_oracle.sh

ORACLE_FILENAME="test_oracle"
FEEDER_ADDRESS=${FEEDER_ADDRESS:-juno1hj5fveer5cjtn4wd6wstzugjfdxzl0xps73ftl}
VALOPER_ADDR=${VALOPER_ADDR:-junovaloper1w20tfhnehc33rgtm9tg8gdtea0svn7twkwj0zq}
LISTEN_ADDR=${LISTEN_ADDR:-0.0.0.0:7171}
JUNO_DIR=$(eval echo "${JUNO_DIR:-"~/.juno1/"}")
echo "$JUNO_DIR"

cd price-feeder
make install

price-feeder version

cp config.example.toml $JUNO_DIR/$ORACLE_FILENAME.toml

# replace gas_price
sed -i 's/0.0001stake/0ujuno/g' $JUNO_DIR/$ORACLE_FILENAME.toml

# replace feeder address
# sed -i 's/address = "juno1w20tfhnehc33rgtm9tg8gdtea0svn7twfnyqee"/address = "juno1hj5fveer5cjtn4wd6wstzugjfdxzl0xps73ftl"/g' $JUNO_DIR/$ORACLE_FILENAME.toml
sed -i 's/address = "juno1w20tfhnehc33rgtm9tg8gdtea0svn7twfnyqee"/address = "'$FEEDER_ADDRESS'"/g' $JUNO_DIR/$ORACLE_FILENAME.toml

# change chain_id
sed -i 's/chain_id = "test-1"/chain_id = "local-1"/g' $JUNO_DIR/$ORACLE_FILENAME.toml
sed -i 's/"chain_id", "test-1"/"chain_id", "local-1"/g' $JUNO_DIR/$ORACLE_FILENAME.toml

# change to running the oracle for the .juno1 directory, so we can get the key
sed -i 's/dir = "\~\/\.juno"/dir = "\~\/\.juno1"/g' $JUNO_DIR/$ORACLE_FILENAME.toml

# VALOPER_ADDR=$(junod q staking validators --node http://localhost:26657 --output json | jq -r '.validators[0].operator_address')

# change validator
sed -i "s/validator = \"junovaloper1w20tfhnehc33rgtm9tg8gdtea0svn7twkwj0zq\"/validator = \"$VALOPER_ADDR\"/g" $JUNO_DIR/$ORACLE_FILENAME.toml

# change 'websocket
sed -i 's/websocket = "stream.binance.com:9443"/websocket = "fstream.binance.com:443"/g' $JUNO_DIR/$ORACLE_FILENAME.toml

# replace listen_addr
sed -i 's/listen_addr = "0.0.0.0:7171"/listen_addr = "'$LISTEN_ADDR'"/g' $JUNO_DIR/$ORACLE_FILENAME.toml

# start it 
price-feeder $JUNO_DIR/$ORACLE_FILENAME.toml