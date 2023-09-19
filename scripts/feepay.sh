# Script to upload a contract to the chain
# Run this in the root of the directly. Cltr + alt + space can be used to run a command in VSCode terminal

CONTRACT_FILE=scripts/fpexample.wasm
JUNOD_NODE=http://localhost:26657

# only run this 1 time per chain start.

junod tx wasm store $CONTRACT_FILE --from juno1 --gas 2000000 -y --chain-id=local-1 --fees=75000ujuno
# Code id from THis is 1 - it is in the above Txhash, just hardcoding since we only need to upload once

sleep 3

# instantiate and get an address
junod tx wasm instantiate 1 '{}' --from juno1 --label "test" --gas 2000000 -y --chain-id=local-1 --fees=75000ujuno --no-admin

sleep 3

# execute on the contract
CONTRACT_ADDR=juno14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9skjuwg8
junod tx wasm execute $CONTRACT_ADDR '{"increment":{}}' --gas 2000000 -y --chain-id=local-1 --fees=75000ujuno --from=juno1

sleep 3

# Query to ensure it went through, else you need to junod q tx <hash> from the above command
junod q wasm contract-state smart $CONTRACT_ADDR '{"get_config":{}}' --chain-id=local-1

sleep 3

# Register Contract
junod tx feepay register juno14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9skjuwg8 3 --home /home/joel/.juno1 --chain-id local-1 --from juno1 --keyring-backend=test --fees=500ujuno -y

sleep 3

# Transfer funds from juno1 to fee pay module
# junod tx bank send juno1 juno1f7f2eapsz6w4s2fytlm34yu5w89ueaesg7x9v5 5000000ujuno --home /home/joel/.juno1 --chain-id local-1 --keyring-backend=test --fees=5000ujuno -y

# sleep 3

# Fund the contract
junod tx feepay fund juno14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9skjuwg8 1000000ujuno --gas=200000 --fees=5000ujuno --home /home/joel/.juno1 --chain-id local-1 --keyring-backend=test -y --from juno1

sleep 3

junod q feepay contract juno14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9skjuwg8 --home /home/joel/.juno1 --chain-id local-1

sleep 3

junod tx feepay update-wallet-limit juno14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9skjuwg8 5 --gas=200000 --fees=5000ujuno --home /home/joel/.juno1 --chain-id local-1 --keyring-backend=test --from juno1 -y

sleep 3 

junod q feepay contract juno14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9skjuwg8 --home /home/joel/.juno1 --chain-id local-1
