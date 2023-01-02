# Unity Upgrade

This upgrade confiscates the funds into the governance-controlled Unity smart contract.

[Unity Upgrade Proposal is viewable here](https://www.mintscan.io/juno/proposals/20).

These are the instructions you will need if you run cosmovisor:

```bash
cd juno
git fetch --tags && git checkout v4.0.0
make build && make install
# this will return commit 299fe4bdee7a7a8b45cd2776359243fdf3630e5a
junod version --long

mkdir -p $DAEMON_HOME/cosmovisor/upgrades/unity/bin && cp $HOME/go/bin/junod $DAEMON_HOME/cosmovisor/upgrades/unity/bin
# this will return v4.0.0
$DAEMON_HOME/cosmovisor/upgrades/unity/bin/junod version
```

Alternatively, you can run the upgrade the old-fashioned way.
