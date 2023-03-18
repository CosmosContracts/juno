# ========================
# === Helper Functions ===
# ========================
function query_contract {
    ARGS=${3:-$JUNOD_COMMAND_ARGS}
    junod query wasm contract-state smart $1 $2 --output json $ARGS
}

function wasm_cmd {
    CONTRACT=$1
    MESSAGE=$2
    FUNDS=$3
    SHOW_LOG=${4:dont_show}
    ARGS=${5:-$JUNOD_COMMAND_ARGS}
    echo "EXECUTE $MESSAGE on $CONTRACT"

    # if length of funds is 0, then no funds are sent
    if [ -z "$FUNDS" ]; then
        FUNDS=""
    else
        FUNDS="--amount $FUNDS"
        echo "FUNDS: $FUNDS"
    fi    

    tx_hash=$(junod tx wasm execute $CONTRACT $MESSAGE $FUNDS $ARGS | jq -r '.txhash')
    export CMD_LOG=$($BINARY query tx $tx_hash --output json | jq -r '.raw_log')    
    if [ "$SHOW_LOG" == "show_log" ]; then
        echo -e "raw_log: $CMD_LOG\n================================\n"
    fi    
}

# CW721
function mint_cw721 {
    CONTRACT_ADDR=$1
    TOKEN_ID=$2
    OWNER=$3
    TOKEN_URI=$4
    EXECUTED_MINT_JSON=`printf '{"mint":{"token_id":"%s","owner":"%s","token_uri":"%s"}}' $TOKEN_ID $OWNER $TOKEN_URI`
    TXMINT=$($BINARY tx wasm execute "$CONTRACT_ADDR" "$EXECUTED_MINT_JSON" $JUNOD_COMMAND_ARGS | jq -r '.txhash') && echo $TXMINT
}