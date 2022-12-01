# Upload a CW20 with an account with no funds as the admin. 
# Then see if transaction fees through the CW20 go to the said account
KEY="juno1"
KEY_ADDR="juno1hj5fveer5cjtn4wd6wstzugjfdxzl0xps73ftl" # test_node.sh
CHAINID="juno-t1"
MONIKER="localjuno"
KEYALGO="secp256k1"
KEYRING="test" # export JUNOD_KEYRING="TEST"
LOGLjunoL="info"
TRACE="" # "--trace"
junod config keyring-backend $KEYRING
junod config chain-id $CHAINID
# junod config output "json"
export JUNOD_NODE="http://localhost:26657"
export JUNOD_COMMAND_ARGS="--gas 5000000 -y --from $KEY --broadcast-mode block --output json --chain-id juno-t1 --fees 5000ujuno"
export JUNOD_COMMANDARGS_FEEACC="--gas 1000000 --gas-prices="0ujuno" -y --from feeacc --broadcast-mode block --output json --chain-id juno-t1"
# junod status

function upload_and_init () {
    ADMIN=$1
    # cw_template = the basic counter contract
    echo "Uploading example contract to chain store"
    cw_template=$(junod tx wasm store cw_template.wasm $JUNOD_COMMAND_ARGS | jq -r '.txhash')
    CWTEMPLATE_CODEID=1
    echo "Instantiating example contract..."    
    CWTEMPLATE_TX_INIT=$(junod tx wasm instantiate "1" '{"count":1}' --label "juno-template" --admin $ADMIN $JUNOD_COMMAND_ARGS -y | jq -r '.txhash') && echo $CWTEMPLATE_TX_INIT
    export CWTEMPLATE_ADDR=$(junod query tx $CWTEMPLATE_TX_INIT --output json | jq -r '.logs[0].events[0].attributes[0].value') && echo "$CWTEMPLATE_ADDR"
}
function balance () {
    ADDRESS=$1
    echo -e "\nBalance of $ADDRESS is:"
    junod q bank balances $ADDRESS
}
function register_fee_share () {
    CONTRACT_ADDR=$1
    ACCOUNT=$2
    # Register for fee share for that given contract
    # junod tx feeshare register $CWTEMPLATE_ADDR juno1efd63aw40lxf3n4mhf7dzhjkr453axurv2zdzk $JUNOD_COMMANDARGS_FEEACC
    echo "Registering $ACCOUNT for fee share for $CONTRACT_ADDR"
    junod tx feeshare register $CONTRACT_ADDR $ACCOUNT $JUNOD_COMMANDARGS_FEEACC
    balance $ACCOUNT
}
function try_to_register_for_non_admin_contract () {    
    # Sets the other account as admin so we can see what happens if we try to register a contract we are not the admin of (fails)    
    echo "THIS WILL FAIL!!! SHOULD RETURN CODE 4 - Setting $KEY_ADDR as admin of $CWTEMPLATE_ADDR"
    upload_and_init juno1hj5fveer5cjtn4wd6wstzugjfdxzl0xps73ftl
    register_fee_share $CWTEMPLATE_ADDR juno1efd63aw40lxf3n4mhf7dzhjkr453axurv2zdzk
}
function execute () {
    CONTRACT_ADDR=$1
    echo "Executing contract 1"
    TX1=$(junod tx wasm execute "$CONTRACT_ADDR" '{"increment":{}}' $JUNOD_COMMAND_ARGS | jq -r '.txhash') && echo $TX1
    echo "Executing contract 2"
    TX2=$(junod tx wasm execute "$CONTRACT_ADDR" '{"reset":{"count":0}}' $JUNOD_COMMAND_ARGS | jq -r '.txhash') && echo $TX2
    balance juno1efd63aw40lxf3n4mhf7dzhjkr453axurv2zdzk
}

upload_and_init juno1efd63aw40lxf3n4mhf7dzhjkr453axurv2zdzk

register_fee_share $CWTEMPLATE_ADDR juno1efd63aw40lxf3n4mhf7dzhjkr453axurv2zdzk
execute $CWTEMPLATE_ADDR

# overrites old contract adderss, should error code 4
try_to_register_for_non_admin_contract

# junod q feeshare contracts
# junod q feeshare contract $CWTEMPLATE_ADDR
# junod q feeshare deployer-contracts juno1efd63aw40lxf3n4mhf7dzhjkr453axurv2zdzk

# TODO: test if you execute 2 messages in 1 Tx on a single contract. Should split fees evenly between each provided they both are registered.
# junod tx wasm execute "$CWTEMPLATE_ADDR" '{"increment":{}}' --from $KEY_ADDR --generate-only | jq  > ~/Desktop/test1.json
# secondMsg=$(cat ~/Desktop/test1.json | jq .body.messages[0])
# # using JQ, append secondMsg to  ~/Desktop/test1.json .body.messages
# cat ~/Desktop/test1.json | jq ".body.messages += [$secondMsg]" > ~/Desktop/test2.json
# # sign it
# junod tx sign ~/Desktop/test2.json --from $KEY --chain-id juno-t1 | jq > ~/Desktop/testsign.json
# junod tx broadcast ~/Desktop/testsign.json

# junod export > ~/Desktop/t.json