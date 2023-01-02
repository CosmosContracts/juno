# Juno Lupercalia Phase 1

Note: _this upgrade takes no direct action on the ongoing CCN issue._

This upgrade moves Juno onto mainline wasmd and includes some security fixes.

The latest version of wasmd will allow the governance module to call smart contracts. This functionality is required before the Unity prop resolution to the CCN issue can be tabled.

For reference, it is likely that the following sequence of upgrades will occur:

1. "lupercalia" upgrade (Lupercalia phase 1/Prop 17)
2. Unity Smart Contract uploaded and instantiated
3. "unity" upgrade (Lupercalia phase 2/Prop 18) - moves funds to Unity SC

[Lupercalia's upgrade Proposal is viewable here](https://www.mintscan.io/juno/proposals/17).

The upgrade is scheduled for block `2582000`, which should be about _17:00 UTC on Tuesday 5th April 2022_.

These are the instructions you will need if you run cosmovisor:

```bash
cd juno
git fetch --tags && git checkout v2.3.0
make build && make install
# this will return commit cfd9b5834bec2bed1ea0fb6a39af787797e4e4ec
junod version --long

mkdir -p $DAEMON_HOME/cosmovisor/upgrades/lupercalia/bin && cp $HOME/go/bin/junod $DAEMON_HOME/cosmovisor/upgrades/lupercalia/bin
# this will return v2.3.0
$DAEMON_HOME/cosmovisor/upgrades/lupercalia/bin/junod version
```

Alternatively, you can run the upgrade the old-fashioned way.

### After the upgrade

To save some space you may also want to delete the old wasm cache at:

```sh
.juno/data/wasm/cache/modules/v1
```

Note: only do this after the upgrade has completed.
