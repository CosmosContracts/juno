# Juno Genesis

For instructions on submitting a Gentx, see the [juno-1 README](./juno-1).

# How to verify Genesis

## Requirements

1. A synced cosmoshub-3 node with pruning default ([backup here](https://archive.interchain.io))
2. Latest [juno](https://github.com/CosmosContracts/Juno) binary on branch `zaki-airdrop`
3. Exchanges list JSON [file](./exchanges.json)

## Parameters

You can find a list of parameters in the following table `parameters.md`

## Procedure

### Init JUNO the genesis file

First we need to clone Juno repository, checkout the mainnet branch `v1.0.0` and compile. Make sure to have go at version `1.16` or above

```
git clone git@github.com:CosmosContracts/juno.git
cd juno
git checkout v1.0.0
make install
```

Delete the old `.juno` directory

**NB** This will delete your validator private key, and all your configurations. So it from a development machine.

```
rm -rf ~/.juno
```

Now we can start forging the genesis file, let's start using the traditional cosmos-sdk init command

```
junod init <moniker> --chain-id juno-1
```

An empty genesis.json file will be created in your `.juno/config` directory.

### Export cosmoshub-3 snapshot

First we need to export state from the hub3 snapshot, if we have a synced node is quite simple with the following command

```
gaiad export --height 5200790 > cosmoshub3.json
```

a file named `cosmoshub3.json` will be generated containing the state of Cosmos Hub on block before the stargate upgrade.

If you don't have a synced node, you can download the export from here [cosmoshub_3_genesis_export.zip](https://gateway.pinata.cloud/ipfs/QmWmFUDFKWfn36De4mTxTJK493EeCE9nh4EtTWt4sUgkp7)

### Generate balances snapshot

Now we need to parse all the balances of cosmoshub 3 and generate a file with filtered accounts, to do so run the following command:

```
junod export-airdrop-snapshot uatom cosmoshub3.json exchanges.json juno_out.json --juno-whalecap 50000000000
```

A new file `juno_out.json` containing all the airdropped balances will be generated.

Some useful statistics will be printed in the command line:

```
cosmos accounts: 78254
atomTotalSupply: 268335648775167
total staked atoms: 179033710700329
extra whale amounts: 118667082985153
total juno airdrop: 30663193590002
```

### Add airdrop accounts & community pool

Now we can add all the airdrop accounts generated in the `juno_out.json` file before, running the following command

```
junod add-airdrop-accounts juno_out.json ujuno 20000000000000
```

It may take from 30 minutes up to a couple of hours, depending on your system specs.

### Add vesting and multisig amounts

To add the core team and multisig vesting accounts we developed an utility to do so [genesis-utils](https://github.com/CosmosContracts/genesis-utils), it's enough to clone that repository and run the following commands:

```
    npm install
    npm run add-vesting-account juno1a8u47ggy964tv9trjxfjcldutau5ls705djqyu fulltime_periods.csv ~/.juno/config/genesis.json
    npm run add-vesting-account juno17py8gfneaam64vt9kaec0fseqwxvkq0flmsmhg fulltime_periods.csv ~/.juno/config/genesis.json
    npm run add-vesting-account juno130mdu9a0etmeuw52qfxk73pn0ga6gawk4k539x partime_periods.csv ~/.juno/config/genesis.json
    npm run add-vesting-account juno18qw9ydpewh405w4lvmuhlg9gtaep79vy2gmtr2 fulltime_periods.csv ~/.juno/config/genesis.json
    npm run add-vesting-account juno1s33zct2zhhaf60x4a90cpe9yquw99jj0zen8pt fulltime_periods.csv ~/.juno/config/genesis.json
    npm run add-vesting-account juno190g5j8aszqhvtg7cprmev8xcxs6csra7xnk3n3 multisig_periods.csv ~/.juno/config/genesis.json
```

### Set Params

```
    npm run set-params ~/.juno/config/genesis.json
```

### Compare SHA256 Sum

Now you can compare the SHA256 hash of the provieded genesis.json with yours.

```
$ jq -S -c -M '' juno-1/pre-genesis.json | shasum -a 256
2341157ce2aec7700b523b2313de19b3203afdf204d555472e29e79fef3e39a1  -


$ jq -S -c -M '' ~/.juno/config/genesis.json | shasum -a 256
2341157ce2aec7700b523b2313de19b3203afdf204d555472e29e79fef3e39a1  -
```
