export JUNOD_NODE="tcp://localhost:26657"

# upload the smart contract, then create a validator. Confirm it works

FLAGS="--from=juno1 --gas=2500000 --fees=50000ujuno --node=http://localhost:26657 --yes --keyring-backend=test --home $HOME/.juno1 --chain-id=local-1 --output=json"

junod tx wasm store ./keeper/contract/juno_staking_hooks_example.wasm $FLAGS

sleep 5

txhash=$(junod tx wasm instantiate 1 '{}' --label=juno_staking --no-admin $FLAGS | jq -r .txhash)
sleep 5
addr=$(junod q tx $txhash --output=json --node=http://localhost:26657 | jq -r .logs[0].events[2].attributes[0].value) && echo $addr

# register addr to staking
junod tx cw-hooks register-staking $addr $FLAGS
junod q cw-hooks staking-contracts

# junod tx cw-hooks unregister-staking $addr $FLAGS
# junod q cw-hooks staking-contracts

# get config
junod q wasm contract-state smart $addr '{"get_config":{}}' --node=http://localhost:26657

# get last validator
junod q wasm contract-state smart $addr '{"last_val_change":{}}' --node=http://localhost:26657
junod q wasm contract-state smart $addr '{"last_delegation_change":{}}' --node=http://localhost:26657

# create validator
junod tx staking create-validator --amount 1ujuno --commission-rate="0.05" --commission-max-rate="1.0" --commission-max-change-rate="1.0" --moniker="test123" --from=juno2 --pubkey=$(junod tendermint show-validator --home $HOME/.juno) --min-self-delegation="1" --gas=1000000 --fees=50000ujuno --node=http://localhost:26657 --yes --keyring-backend=test --home $HOME/.juno1 --chain-id=local-1 --output=json
