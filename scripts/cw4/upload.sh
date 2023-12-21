JUNOD_NODE=http://127.0.0.1:26657

FLAGS="--chain-id=local-1 --gas=auto --gas-adjustment=2.0 --fees=100000ujuno --node=$JUNOD_NODE --from=juno1 --keyring-backend=test --home=$HOME/.juno1 --yes"

junod tx wasm store ./scripts/cw4/cw4_group.wasm $FLAGS

junod tx wasm instantiate 1 '{"admin":"juno1hj5fveer5cjtn4wd6wstzugjfdxzl0xps73ftl","members":[{"addr":"juno130mdu9a0etmeuw52qfxk73pn0ga6gawk4k539x","weight":1},{"addr":"juno17py8gfneaam64vt9kaec0fseqwxvkq0flmsmhg","weight":1},{"addr":"juno18qw9ydpewh405w4lvmuhlg9gtaep79vy2gmtr2","weight":1},{"addr":"juno1ra4mme6sr5r6prqhzan03mz03jez6s2twplwmd","weight":1},{"addr":"juno1s33zct2zhhaf60x4a90cpe9yquw99jj0zen8pt","weight":1},{"addr":"juno1u93z4xlptl9ujx6pq3y0w4phdkj5mapgk0wuuj","weight":1}]}' $FLAGS --no-admin --label="cw4-core1"

CONTRACT=juno14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9skjuwg8

junod q wasm contract-state smart $CONTRACT '{"list_members":{}}'


junod tx wasm execute $CONTRACT '{"update_members":{"remove":["juno17py8gfneaam64vt9kaec0fseqwxvkq0flmsmhg"],"add":[]}}' $FLAGS