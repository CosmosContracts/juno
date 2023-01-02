# Juno v10 Upgrade

Juno v10 brings two important changes to Juno:

- Changes to the mint module
- Changes to the ICA configuration of the chain in order to enable liquid staking chains over ICA

Additional links:

- [The upgrade prop is viewable here](https://www.mintscan.io/juno/proposals/40).
- [The v10.0.2 changelog is viewable here](https://github.com/CosmosContracts/juno/releases/tag/v10.0.2).

The target block for this upgrade is [5004269](https://www.mintscan.io/juno/blocks/5004269), which is expected to arrive on _Wed Sep 28 2022 at 1400UTC_, +/- 1 hour.

If you want to confirm the changes made between `v10.0.0` and `v10.0.2`, you can see them in the upgrade handler in `app.go` on L852 and L865. These two messages were missed in the first RC for Juno v10.

These are the instructions you will need if you run cosmovisor:

```bash
cd juno
git fetch --tags && git checkout v10.0.2
make build && make install
# this will return commit f2f9de4467be8ce64f86eded128b87c9364fd39a
junod version --long

mkdir -p $DAEMON_HOME/cosmovisor/upgrades/v10/bin && cp $HOME/go/bin/junod $DAEMON_HOME/cosmovisor/upgrades/v10/bin
# this will return v10.0.2
$DAEMON_HOME/cosmovisor/upgrades/v10/bin/junod version
```

Alternatively, you can run the upgrade the old-fashioned way.
