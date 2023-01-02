# Juno v11 Upgrade

Juno v11 is a minor change that adds:

- bug fixes
- new ICA configuration
- a pruning command
- CosmWasm 1.1

**Important!** you _MUST_ apply this config in `app.toml` when upgrading. This should go in the base configuration section, under  the `index-events = []` configuration line:

```toml
# IavlCacheSize set the size of the iavl tree cache.
# Default cache size is 50mb.
iavl-cache-size = 781250

# IAVLDisableFastNode enables or disables the fast node feature of IAVL.
# Default is true.
iavl-disable-fastnode = false
```

Additional links:

- [The upgrade prop is viewable here](https://www.mintscan.io/juno/proposals/47).
- [The v11.0.0 changelog is viewable here](https://github.com/CosmosContracts/juno/releases/tag/v11.0.0).

The target block for this upgrade is [5480000](https://www.mintscan.io/juno/blocks/5480000), which is expected to arrive on _Mon Oct 31 2022 at 1700UTC_, +/- 1 hour.

These are the instructions you will need if you run cosmovisor:

```bash
cd juno
git fetch --tags && git checkout v11.0.0
make build && make install
# this will return commit b27fc7d9312267b293d3355dd4a06523d76e247f
junod version --long

mkdir -p $DAEMON_HOME/cosmovisor/upgrades/v11/bin && cp $HOME/go/bin/junod $DAEMON_HOME/cosmovisor/upgrades/v11/bin
# this will return v11.0.0
$DAEMON_HOME/cosmovisor/upgrades/v11/bin/junod version
```

Alternatively, you can run the upgrade the old-fashioned way.
