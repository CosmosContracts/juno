junod tx wasm store oracle_querier.wasm --from mykey --chain-id test-1 --gas 10000000 --fees 20000stake
junod tx wasm instantiate 4 "{}" --from mykey --label "BRO CW20" -y  --gas 10000000 --fees 10000stake --no-admin   
junod q wasm contract-state smart juno14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9skjuwg8

junod q wasm contract-state smart juno1ghd753shjuwexxywmgs4xz7x2q732vcnkm6h2pyv9s6ah3hylvrq722sry '{"exchange_rate": {"denom":"stake"}}'