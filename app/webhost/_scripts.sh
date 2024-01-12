# EXECUTE THIS SCRIPT FROM THIS DIRECTORY

junod tx wasm store cw_webhost.wasm --from=juno1 --keyring-backend=test --chain-id=local-1 --gas=auto --gas-adjustment=1.5 --gas-prices=0.025ujuno --yes --home=$HOME/.juno1

sleep 3

junod tx wasm instantiate 1 '{}' --from=juno1 --keyring-backend=test --chain-id=local-1 --label=cw_webhost --gas=auto --gas-adjustment=1.5 --gas-prices=0.025ujuno --yes --no-admin --home=$HOME/.juno1
ADDRESS=juno14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9skjuwg8

sleep 3

# Encode zip file to base64, generate new_web_cmd.txt
# ENSURE YOU HAVE A FILE NAMED 'web.zip' in the same directory as this script
python3 encode_zip_to_base64.py
JSON="$(cat new_web_cmd.txt)"
junod tx wasm execute $ADDRESS $JSON --from=juno1 --keyring-backend=test --chain-id=local-1 --gas=auto --gas-adjustment=1.5 --gas-prices=0.025ujuno --yes  --home=/home/joel/.juno1

sleep 3

junod q wasm contract-state smart $ADDRESS '{"get_website":{"name":"test"}}' --home=$HOME/.juno1

# Website should be available at: http://localhost:1317/webhost/juno14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9skjuwg8/test/