juno-local

junod tx wasm store /home/reece/Desktop/cw-webhost/artifacts/cw_webhost.wasm --from=juno1 --keyring-backend=test --chain-id=local-1 --gas=auto --gas-adjustment=1.5 --gas-prices=0.025ujuno --yes

junod tx wasm instantiate 1 '{}' --from=juno1 --keyring-backend=test --chain-id=local-1 --label=cw_webhost --gas=auto --gas-adjustment=1.5 --gas-prices=0.025ujuno --yes --no-admin

ADDRESS=juno14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9skjuwg8


junod tx wasm execute $ADDRESS '{"new_website":{"name":"test","source":"<html><script>alert(\"popup\")</script><style>body {background-color: lightblue;}</style><h1>Test Website Header</h1><p>Paragraph</p></html>"}}' --from=juno1 --keyring-backend=test --chain-id=local-1 --gas=auto --gas-adjustment=1.5 --gas-prices=0.025ujuno --yes

junod q wasm contract-state smart $ADDRESS '{"get_website":{"name":"test"}}'