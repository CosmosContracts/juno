# Moneta Upgrade

The time has come, folks. It's time to get CosmWasm on mainnet and light this bottle rocket 🚀

[You can see the Moneta Governance Proposal here](https://www.mintscan.io/juno/proposals/8).

The upgrade is scheduled for block `1055000`, which should be about _16:30PM UTC on 15th December 2021_.

There's a dashboard from `cros-nest` that allows you to [check the rough ETA](https://chain-monitor.cros-nest.com/d/Upgrades/upgrades?var-chain_id=juno-1&orgId=1&refresh=1m), but if in doubt, check back on Discord.

Nevertheless, you should really be automating this rather than switching over manually. Most mainnet validators are running Cosmovisor, and [a setup guide can be found in our docs](https://docs.junonetwork.io/validators/setting-up-cosmovisor).

If you use cosmovisor, do this now:

```bash
# get the new version (from inside the juno repo)
git checkout main && git pull
git checkout v2.0.6
make build && make install

# check the version - should be v2.0.6
# junod version --long will be commit d9c8ee6d13076f549688662aaeade4499e108d15
junod version

# make a directory if you haven't already
mkdir -p $DAEMON_HOME/cosmovisor/upgrades/moneta/bin

# if you are using cosmovisor you then need to copy this new binary
cp /home/<your-user>/go/bin/junod $DAEMON_HOME/cosmovisor/upgrades/moneta/bin

# find out what version you are about to run - should be v2.0.6
$DAEMON_HOME/cosmovisor/upgrades/moneta/bin/junod version
```

When the block height is reached, Cosmovisor will backup the chain data and point to the new v2.0.6 binary

## Fees

Fees are set by each individual validator. The ability to spam TXs (Smart Contracts or otherwise) is a potential attack vector for the Juno network, so if you haven't already set minimum gas prices, please change these in `.juno/config/app.toml`. We suggest `0.025ujuno`.

```
sed -i.bak -e "s/^minimum-gas-prices *=.*/minimum-gas-prices = \"0.025ujuno\"/" ~/.juno/config/app.toml
```

## Commission

[Juno Governance Proposal 3](https://www.mintscan.io/juno/proposals/3) signalled the community wanted to impose a minimum validator commission. If your validator has already increased to the minimum, then thank you. All other validators will have their commission forcibly increased by this update, if it is below the minimum.

## A note on this upgrade

Although we have tried our best to test this fully (thanks to validators who participated in `uni`, and extra special thanks to those that participated in `astarte`, `astarte-1` and `astarte-2`), we are running CosmWasm in a permissionless state on a public blockchain.

As a result, there may be security or functional patches that are needed simply because mainnet is impossible to replicate on a testnet.

This means that for a period of time after the Moneta release, we would expect the validator community to be extra vigilant for such changes - check back in on Discord to see what's happening in the validator channel.

If you find an issue, report it in Discord. If it's a security issue, then contact a core dev as per [the Juno security disclosure guidance](https://github.com/CosmosContracts/juno/blob/main/SECURITY.md).

## Closing thoughts

WAGMI
