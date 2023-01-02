# Veritas Upgrade

This upgrade relocates the confiscated funds to the governance-controlled Unity smart contract.

[Veritas Upgrade Proposal is viewable here](https://www.mintscan.io/juno/proposals/21).

These are the instructions you will need if you run cosmovisor:

```bash
cd juno
git fetch --tags && git checkout v5.0.1
make build && make install
# this will return commit 77fdc2e9b0b380f640286745356384c64d86fd32
junod version --long

mkdir -p $DAEMON_HOME/cosmovisor/upgrades/veritas/bin && cp $HOME/go/bin/junod $DAEMON_HOME/cosmovisor/upgrades/veritas/bin
# this will return v5.0.1
$DAEMON_HOME/cosmovisor/upgrades/veritas/bin/junod version
```

Alternatively, you can run the upgrade the old-fashioned way.
