# Junø

![c11](https://user-images.githubusercontent.com/79812965/131373443-5ff0d9f6-2e2a-41bd-8347-22ac4983e625.jpg)

❗️For issue disclosure, check out [SECURITY.md](./SECURITY.md) ❗️

🚀 For release procedures, check out [RELEASES.md](./RELEASES.md). 🚀

**Juno** is an open-source platform for inter-operable smart contracts which automatically execute, control or document
a procedure of events and actions according to the terms of the contract or agreement to be valid and usable across
multiple sovereign networks.

Juno is a **sovereign public blockchain** in the Cosmos ecosystem. It aims to provide a sandbox environment for the
deployment of inter-operable smart contracts. The network serves as a **decentralized, permissionless**, and *
*censorship-resistant** zone for developers to efficiently and securely launch application-specific smart contracts.

Juno originates from a **community-driven initiative**, prompted by developers, validators and delegators in the Cosmos
ecosystem. The common vision is to preserve the neutrality, performance, and reliability of the Cosmos Hub. This is
achieved by offloading smart contract deployment to a dedicated sister Hub.

**Juno** is a blockchain built using Cosmos SDK and Tendermint.

## Get started

If you have [Docker](https://www.docker.com/) installed, then you can run a local node with a single command.

```bash
STAKE_TOKEN=ujunox UNSAFE_CORS=true TIMEOUT_COMMIT=1s docker-compose up
```

## Learn more

- [Juno](https://junonetwork.io)
- [Discord](https://discord.gg/QcWPfK4gJ2)
- [Telegram](https://t.me/JunoNetwork)
- [Cosmos SDK documentation](https://docs.cosmos.network)
- [Cosmos SDK Tutorials](https://tutorials.cosmos.network)

## Attribution

We'd like to thank the following teams for their contributions to Juno:

- [EVMOS](https://twitter.com/EvmosOrg) - x/feeshare
- [tgrade](https://twitter.com/TgradeFinance) - x/globalfee
- [confio](https://twitter.com/confio_tech) - CosmWasm
- [osmosis](https://twitter.com/osmosiszone) - Osmosis

## My docs

### Run juno chain

Rm old config

```
lsof -i :26657
kill -9 21640
./bin/junod reset app
./bin/junod reset wasm
rm -rf ~/.juno
```

Create old config

```
./bin/junod init my-node --chain-id=my-chain
```

Create validator

```
./bin/junod keys add my-validator --keyring-backend=test
```

Output:

```
- address: juno1fgt6akzfp7qls5qctmpm4n0pvfu43dvvqekz60
  name: my-validator
  pubkey: '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AzS1YVbmwglWlnyl62W6twJhEOXIcKMzOfw8wfcl6s+/"}'
  type: local
```

Funding balance for validator

```
./bin/junod genesis add-genesis-account $(./bin/junod keys show my-validator -a --keyring-backend=test) 100000000stake
```

Configure validator in genesis

```
./bin/junod genesis gentx my-validator 1000000stake \
  --chain-id=my-chain \
  --keyring-backend=test \
  --moniker="MyValidator" \
  --commission-rate="0.10" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" \
  --min-self-delegation="1"
```

Create my-wallet

```
./bin/junod keys add my-wallet --keyring-backend=test
```

Output:

```
- address: juno1e3rdxdlp9zdskp3d4p03yl7ae728mz0gusrvyj
  name: my-wallet
  pubkey: '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"A+ixPL54+EYrd0HAr8TBy7i5fpf+C93itlTNJAvisPvx"}'
  type: local
```

Funding balance for my-wallet

```
./bin/junod genesis add-genesis-account $(./bin/junod keys show my-wallet -a --keyring-backend=test) 1000000000stake
```

Check

```
./bin/junod genesis collect-gentxs
./bin/junod genesis validate-genesis
```

Configure min gas price in app.toml and run juno

```
./bin/junod start
./bin/junod start --minimum-gas-prices=0.025stake
```

### Test tokenfactory

Check balance for my-wallet

```
./bin/junod query bank balances $(./bin/junod keys show my-wallet -a --keyring-backend=test)
```

Check tokenfactory

```
./bin/junod query tokenfactory params
```

Output:

```
params:
  denom_creation_fee:
  - amount: "10000000"
    denom: stake
  denom_creation_gas_consume: "2000000"
```

Create mytoken

```
./bin/junod tx tokenfactory create-denom mytoken --from=my-wallet --chain-id=my-chain --keyring-backend=test --gas=auto --gas-adjustment=1.5 --fees=80000stake -y
```

Output:

```
gas estimate: 3153708
code: 0
codespace: ""
data: ""
events: []
gas_used: "0"
gas_wanted: "0"
height: "0"
info: ""
logs: []
raw_log: '[]'
timestamp: ""
tx: null
txhash: FDC0D05FDFFD46F82FA2201A1A31DB90BE23AA801D384C5A1E22755C3E17886B
```

Check mytoken

```
./bin/junod query tokenfactory denoms-from-creator $(./bin/junod keys show my-wallet -a --keyring-backend=test)
```

Output:

```
denoms:
- factory/juno1e56qnzv38pdrlkqtwfkkx5cmugrw76t55thjhe/mytoken
```

Check mytoken metadata

```
./bin/junod query tokenfactory denom-authority-metadata factory/juno1e56qnzv38pdrlkqtwfkkx5cmugrw76t55thjhe/mytoken
```

Output:

```
authority_metadata:
admin: juno1e56qnzv38pdrlkqtwfkkx5cmugrw76t55thjhe
```

Mint mytoken

```
./bin/junod tx tokenfactory mint 1000factory/juno1e56qnzv38pdrlkqtwfkkx5cmugrw76t55thjhe/mytoken --from=my-wallet --chain-id=my-chain --keyring-backend=test --fees=5000stake -y
```

Mint mytoken with short amount format

```
./bin/junod tx tokenfactory mint 1000mytoken --from=my-wallet --chain-id=my-chain --keyring-backend=test --fees=5000stake -y
```

Output:

```
code: 0
codespace: ""
data: ""
events: []
gas_used: "0"
gas_wanted: "0"
height: "0"
info: ""
logs: []
raw_log: '[]'
timestamp: ""
tx: null
txhash: EAD728AE54DAEB071D153EA04DA3C4B71F4091ADE413CC12AA6E4A74F8FAFBBF
```

Check balance for my-wallet

```
/bin/junod query bank balances $(./bin/junod keys show my-wallet -a --keyring-backend=test)
```

Output:

```
balances:
- amount: "1000"
  denom: factory/juno1e56qnzv38pdrlkqtwfkkx5cmugrw76t55thjhe/mytoken
- amount: "989805000"
  denom: stake
pagination:
  next_key: null
  total: "0"
```

Check tx

```
./bin/junod query tx EAD728AE54DAEB071D153EA04DA3C4B71F4091ADE413CC12AA6E4A74F8FAFBBF
```

Create recipient-wallet

```
./bin/junod keys add recipient-wallet --keyring-backend=test
```

Output:

```
- address: juno1abcdefg1234567890hijklmnopqrstuvwxy
  name: recipient-wallet
  pubkey: '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"..."}'
  type: local
```

Funding balance for recipient-wallet

```
./bin/junod tx bank send my-wallet $(./bin/junod keys show recipient-wallet -a --keyring-backend=test) 100000stake --chain-id=my-chain --keyring-backend=test --fees=5000stake -y
```

Output:

```
code: 0
codespace: ""
data: ""
events: []
gas_used: "0"
gas_wanted: "0"
height: "0"
info: ""
logs: []
raw_log: '[]'
timestamp: ""
tx: null
txhash: 0BBF2F9DC992A85E772AD693AC7649540B377B2146945DBFBE3EAC05F8D9C0DA
```

Check balance for recipient-wallet

```
./bin/junod query bank balances $(./bin/junod keys show recipient-wallet -a --keyring-backend=test)
```

Output:

```
balances:
- amount: "100000"
  denom: stake
pagination:
  next_key: null
  total: "0"
```

Send mytoken from my-wallet to recipient-wallet:

```
./bin/junod tx bank send my-wallet $(./bin/junod keys show recipient-wallet -a --keyring-backend=test) 100factory/juno1e56qnzv38pdrlkqtwfkkx5cmugrw76t55thjhe/mytoken --chain-id=my-chain --keyring-backend=test --fees=5000stake -y
```

Output:

```
code: 0
codespace: ""
data: ""
events: []
gas_used: "0"
gas_wanted: "0"
height: "0"
info: ""
logs: []
raw_log: '[]'
timestamp: ""
tx: null
txhash: 2E7435F7A30AEF3F83E165FB1F8F387B2E4011BD98B6A8E9C73339609470493C
```

Check balance for recipient-wallet

```
./bin/junod query bank balances $(./bin/junod keys show recipient-wallet -a --keyring-backend=test)
```

Output:

```
balances:
- amount: "100"
  denom: factory/juno1e56qnzv38pdrlkqtwfkkx5cmugrw76t55thjhe/mytoken
- amount: "100000"
  denom: stake
pagination:
  next_key: null
  total: "0"
```

### Test Rust smart-contract

Install and configure Rust:

```
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh

rustup target add wasm32-unknown-unknown

rustup update stable
```

Build

```
cd smartcontracts/hello_world
cargo build --release --target wasm32-unknown-unknown
```

Optimize with docker

```
docker run --rm -v "$(pwd)":/code \
  --mount type=volume,source="$(basename "$(pwd)")_cache",target=/target \
  --mount type=volume,source=registry_cache,target=/usr/local/cargo/registry \
  cosmwasm/optimizer:0.16.1
```

Upload contract to juno chain

```
cd ../..
./bin/junod tx wasm store smartcontracts/hello_world/artifacts/hello_world.wasm --from my-wallet --chain-id my-chain --gas=auto --fees=500stake -y
```

Instantiate the Contract

```
junod tx wasm instantiate <code_id> '{}' --from my-wallet --label "HelloWorld" --no-admin --chain-id my-chain --gas=auto --fees=500stake -y
```

Interact with the Contract
```
junod query wasm contract-state smart <contract_address> '{}'
```

Output:
```
"Hello, World!"
```
