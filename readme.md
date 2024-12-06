# Junø

![c11](https://user-images.githubusercontent.com/79812965/131373443-5ff0d9f6-2e2a-41bd-8347-22ac4983e625.jpg)

❗️For issue disclosure, check out [SECURITY.md](./SECURITY.md) ❗️

🚀 For release procedures, check out [RELEASES.md](./RELEASES.md). 🚀

**Juno** is an open-source platform for inter-operable smart contracts which automatically execute, control or document a procedure of events and actions according to the terms of the contract or agreement to be valid and usable across multiple sovereign networks.

Juno is a **sovereign public blockchain** in the Cosmos ecosystem. It aims to provide a sandbox environment for the deployment of inter-operable smart contracts. The network serves as a **decentralized, permissionless**, and **censorship-resistant** zone for developers to efficiently and securely launch application-specific smart contracts.

Juno originates from a **community-driven initiative**, prompted by developers, validators and delegators in the Cosmos ecosystem. The common vision is to preserve the neutrality, performance, and reliability of the Cosmos Hub. This is achieved by offloading smart contract deployment to a dedicated sister Hub.

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


## Моя дока

### Базовый запуск juno

Удаляем прошлую конфигурацию
```
rm -rf ~/.juno
```

Создаем новую конфигурацию
```
./bin/junod init my-node --chain-id=my-chain
```

Создаем validator
```
./bin/junod keys add my-validator --keyring-backend=test
```
Вывод:
```
- address: juno1fgt6akzfp7qls5qctmpm4n0pvfu43dvvqekz60
  name: my-validator
  pubkey: '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AzS1YVbmwglWlnyl62W6twJhEOXIcKMzOfw8wfcl6s+/"}'
  type: local
```

Пополняем баланс для validator
```
./bin/junod genesis add-genesis-account $(./bin/junod keys show my-validator -a --keyring-backend=test) 100000000stake
```
Настраиваем validator в genesis
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

Создаем кошелек
```
./bin/junod keys add my-wallet --keyring-backend=test
```
Вывод:
```
- address: juno1e3rdxdlp9zdskp3d4p03yl7ae728mz0gusrvyj
  name: my-wallet
  pubkey: '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"A+ixPL54+EYrd0HAr8TBy7i5fpf+C93itlTNJAvisPvx"}'
  type: local
```

Пополняем баланс для кошелька
```
./bin/junod genesis add-genesis-account $(./bin/junod keys show my-wallet -a --keyring-backend=test) 1000000000stake
```

Проверяем, что все корректно
```
./bin/junod genesis collect-gentxs
./bin/junod genesis validate-genesis
```

Нужно задать min gas price в app.toml или запустить с доп флагом
```
./bin/junod start
./bin/junod start --minimum-gas-prices=0.025stake
```

В app.toml выставить такие значения
```
[api]

# Enable defines if the API server should be enabled.
enable = true

# Swagger defines if swagger documentation should automatically be registered.
swagger = false

# Address defines the API server to listen on.
address = "tcp://0.0.0.0:1317"

# MaxOpenConnections defines the number of maximum open connections.
max-open-connections = 1000

# RPCReadTimeout defines the Tendermint RPC read timeout (in seconds).
rpc-read-timeout = 10

# RPCWriteTimeout defines the Tendermint RPC write timeout (in seconds).
rpc-write-timeout = 0

# RPCMaxBodyBytes defines the Tendermint maximum request body (in bytes).
rpc-max-body-bytes = 1000000

# EnableUnsafeCORS defines if CORS should be enabled (unsafe - use it at your own risk).
enabled-unsafe-cors = true
```

Проверить, что api работает:
```
curl http://localhost:1317/cosmos/tokenfactory/v1beta1/params
```
Вывод:
```
{"code":12,"message":"Not Implemented","details":[]}
```

Запросить статус:
```
curl http://localhost:26657/status
```

### Тестируем tokenfactory

Проверяем, что баланс успешно добавлен для кошелька my-wallet
```
./bin/junod query bank balances $(./bin/junod keys show my-wallet -a --keyring-backend=test)
```

Проверяем, что tokenfactory правильно настроен
```
./bin/junod query tokenfactory params
```
Вывод:
```
params:
  denom_creation_fee:
  - amount: "10000000"
    denom: stake
  denom_creation_gas_consume: "2000000"
```

Создаем новый токен через tokenfactory
```
./bin/junod tx tokenfactory create-denom mytoken --from=my-wallet --chain-id=my-chain --keyring-backend=test --gas=auto --gas-adjustment=1.5 --fees=80000stake -y
```
Вывод:
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

Проверяем созданный токен
```
./bin/junod query tokenfactory denoms-from-creator $(./bin/junod keys show my-wallet -a --keyring-backend=test)
```
Вывод:
```
denoms:
- factory/juno1e56qnzv38pdrlkqtwfkkx5cmugrw76t55thjhe/mytoken
```

Проверяем метаданные токена
```
./bin/junod query tokenfactory denom-authority-metadata factory/juno1e56qnzv38pdrlkqtwfkkx5cmugrw76t55thjhe/mytoken
```
Вывод:
```
authority_metadata:
admin: juno1e56qnzv38pdrlkqtwfkkx5cmugrw76t55thjhe
```

Делаем mint для созданного токена
```
./bin/junod tx tokenfactory mint 1000factory/juno1e56qnzv38pdrlkqtwfkkx5cmugrw76t55thjhe/mytoken --from=my-wallet --chain-id=my-chain --keyring-backend=test --fees=5000stake -y
```
Вывод:
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

Проверяем баланс для кошелька my-wallet:
```
/bin/junod query bank balances $(./bin/junod keys show my-wallet -a --keyring-backend=test)
```
Вывод:
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

Проверяем транзакцию:
```
./bin/junod query tx EAD728AE54DAEB071D153EA04DA3C4B71F4091ADE413CC12AA6E4A74F8FAFBBF
```


Создаем новый кошелек recipient-wallet:
```
./bin/junod keys add recipient-wallet --keyring-backend=test
```
Вывод:
```
- address: juno1abcdefg1234567890hijklmnopqrstuvwxy
  name: recipient-wallet
  pubkey: '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"..."}'
  type: local
```

Пополняем баланс для recipient-wallet:
```
./bin/junod tx bank send my-wallet $(./bin/junod keys show recipient-wallet -a --keyring-backend=test) 100000stake --chain-id=my-chain --keyring-backend=test --fees=5000stake -y
```
Вывод:
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

Проверяем баланс для recipient-wallet:
```
./bin/junod query bank balances $(./bin/junod keys show recipient-wallet -a --keyring-backend=test)
```
Вывод:
```
balances:
- amount: "100000"
  denom: stake
pagination:
  next_key: null
  total: "0"
```

Переводим новый токен с my-wallet на recipient-wallet:
```
./bin/junod tx bank send my-wallet $(./bin/junod keys show recipient-wallet -a --keyring-backend=test) 100factory/juno1e56qnzv38pdrlkqtwfkkx5cmugrw76t55thjhe/mytoken --chain-id=my-chain --keyring-backend=test --fees=5000stake -y
```
Вывод:
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

Проверяем баланс для recipient-wallet:
```
./bin/junod query bank balances $(./bin/junod keys show recipient-wallet -a --keyring-backend=test)
```
Вывод:
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
