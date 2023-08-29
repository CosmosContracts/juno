# Script to upload a contract to the chain
# Run this in the root of the directly. Cltr + alt + space can be used to run a command in VSCode terminal

CONTRACT_FILE=scripts/fpexample.wasm
JUNOD_NODE=http://localhost:26657

# only run this 1 time per chain start.

junod tx wasm store $CONTRACT_FILE --from juno1 --gas 2000000 -y --chain-id=local-1 --fees=75000ujuno
# Code id from THis is 1 - it is in the above Txhash, just hardcoding since we only need to upload once

# instantiate and get an address
junod tx wasm instantiate 1 '{}' --from juno1 --label "test" --gas 2000000 -y --chain-id=local-1 --fees=75000ujuno --no-admin

# execute on the contract
CONTRACT_ADDR=juno14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9skjuwg8
junod tx wasm execute $CONTRACT_ADDR '{"increment":{}}' --gas 2000000 -y --chain-id=local-1 --fees=75000ujuno --from=juno1

# Query to ensure it went through, else you need to junod q tx <hash> from the above command
junod q wasm contract-state smart $CONTRACT_ADDR '{"get_config":{}}' --chain-id=local-1