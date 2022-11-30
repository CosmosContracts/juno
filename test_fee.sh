# Upload a CW20 with an account with no funds as the admin. 
# Then see if transaction fees through the CW20 go to the said account

KEY="juno1"
KEY_ADDR="juno1hj5fveer5cjtn4wd6wstzugjfdxzl0xps73ftl" # test_node.sh
CHAINID="juno-t1"
MONIKER="localjuno"
KEYALGO="secp256k1"
KEYRING="test" # export juno_KEYRING="TEST"
LOGLjunoL="info"
TRACE="" # "--trace"
junod config keyring-backend $KEYRING
junod config chain-id $CHAINID
# junod config output "json"
export JUNOD_NODE="http://localhost:26657"
export JUNOD_COMMAND_ARGS="--gas 5000000 -y --from $KEY --broadcast-mode block --output json --chain-id juno-t1 --fees 500ujuno"
export JUNOD_COMMANDARGS_FEEACC="--gas 1000000 --gas-prices="0ujuno" -y --from feeacc --broadcast-mode block --output json --chain-id juno-t1"
# junod status

# cw_template = the basic counter contract
cw_template=$(junod tx wasm store cw_template.wasm $JUNOD_COMMAND_ARGS | jq -r '.txhash')
CWTEMPLATE_TX_INIT=$(junod tx wasm instantiate "1" '{"count":1}' --label "juno-template" --admin juno1efd63aw40lxf3n4mhf7dzhjkr453axurv2zdzk $JUNOD_COMMAND_ARGS -y | jq -r '.txhash') && echo $CWTEMPLATE_TX_INIT
CWTEMPLATE_CODEID=1

CWTEMPLATE_ADDR=$(junod query tx $CWTEMPLATE_TX_INIT --output json | jq -r '.logs[0].events[0].attributes[0].value') && echo "$CWTEMPLATE_ADDR"
TX1=$(junod tx wasm execute "$CWTEMPLATE_ADDR" '{"increment":{}}' $JUNOD_COMMAND_ARGS | jq -r '.txhash') && echo $TX1
TX2=$(junod tx wasm execute "$CWTEMPLATE_ADDR" '{"reset":{"count":0}}' $JUNOD_COMMAND_ARGS | jq -r '.txhash') && echo $TX2

junod q wasm contract-state smart "$CWTEMPLATE_ADDR" '{"get_count":{}}' | jq -r '.data.count'
junod tx feeshare register $CWTEMPLATE_ADDR juno1efd63aw40lxf3n4mhf7dzhjkr453axurv2zdzk $JUNOD_COMMANDARGS_FEEACC

# This needs to check the from address is in the wasm contract info as admin.
junod tx feeshare register $CWTEMPLATE_ADDR juno1efd63aw40lxf3n4mhf7dzhjkr453axurv2zdzk $JUNOD_COMMANDARGS_FEEACC


junod tx feeshare update $CWTEMPLATE_ADDR juno1efd63aw40lxf3n4mhf7dzhjkr453axurv2zdzk $JUNOD_COMMANDARGS_FEEACC 

# junod q feeshare contracts
junod q feeshare contract $CWTEMPLATE_ADDR

junod q feeshare deployer-contracts juno1efd63aw40lxf3n4mhf7dzhjkr453axurv2zdzk



# === Actual logic time ===

function sendCw20Msg() {
    BASE64_MSG=$(echo -n "{"receive":{}}" | base64)
    export EXECUTED_MINT_JSON=`printf '{"send":{"contract":"%s","amount":"%s","msg":"%s"}}' $BURN_ADDR "5" $BASE64_MSG`

    TX=$(junod tx wasm execute "$CW20_ADDR" "$EXECUTED_MINT_JSON" $JUNOD_COMMAND_ARGS | jq -r '.txhash') && echo $TX    
    # junod tx wasm execute "$CW20_ADDR" `printf '{"send":{"contract":"%s","amount":"5","msg":"e3JlZGVlbTp7fX0="}}' $BURN_ADDR` $JUNOD_COMMAND_ARGS
}
junod tx wasm execute $CW20_ADDR '{"send":{"contract":"","amount":"100"}}' $JUNOD_COMMAND_ARGS


# # junod tx feeshare register $CW20_ADDR juno1efd63aw40lxf3n4mhf7dzhjkr453axurv2zdzk $JUNOD_COMMANDARGS_FEEACC 

# sendCw20Msg

junod q bank balances juno1efd63aw40lxf3n4mhf7dzhjkr453axurv2zdzk