# Moneta Patch

This is a patch that covers issues discovered after the mainnet Moneta launch.

The upgrade is scheduled for block `1165200`, which should be about _15:30PM UTC on 23rd December 2021_.

There's a dashboard from `cros-nest` that allows you to [check the rough ETA](https://chain-monitor.cros-nest.com/d/Upgrades/upgrades?orgId=1&refresh=1m), but if in doubt, check back on Discord.

Here's the instructions for using Cosmovisor. Most mainnet validators are running Cosmovisor, and [a setup guide can be found in our docs](https://docs.junonetwork.io/validators/setting-up-cosmovisor).

```bash
# get the new version (run inside the repo)
git checkout main && git pull
git checkout v2.1.0
make build && make install

# check the version - should be v2.1.0
# junod version --long will be commit e6b8c212b178cf575035065b78309aed547b1335
junod version

# make a dir if you've not
mkdir -p $DAEMON_HOME/cosmovisor/upgrades/moneta-patch/bin

# if you are using cosmovisor you then need to copy this new binary
cp /home/<your-user>/go/bin/junod $DAEMON_HOME/cosmovisor/upgrades/moneta-patch/bin

# find out what version you are about to run - should be v2.1.0
$DAEMON_HOME/cosmovisor/upgrades/moneta-patch/bin/junod version
```

If you are not using cosmovisor, then the chain will halt at the target height and you can manually switch over.

## A note on this upgrade

We had hoped to move straight to `v2.2.0`, which will contain a patch for a recently-disclosed security issue. However, after putting together a patch and testing, the team were not happy with the risk involved, so it will instead go through a proper testnet cycle to establish that it is solid before release.

We are sorry about the timing and notice involved in shipping this patch, and the follow-up.

If you find an issue, report it in Discord. If it's a security issue, then contact a core dev as per [the Juno security disclosure guidance](https://github.com/CosmosContracts/juno/blob/main/SECURITY.md).
