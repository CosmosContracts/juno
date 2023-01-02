# Multiverse Upgrade

**NB:** This upgrade was originally using the `v7.0.0` binary, but ultimately the `v8.0.0` binary had to be used to get around [this issue](https://github.com/cosmos/cosmos-sdk/issues/11707).

Multiverse adds ICA Host functionality to the Juno chain, and exposes the following actions to Controller Chains:

- Native functionality and core SDK modules
- Smart Contract store
- Smart Contract instantiate
- Smart Contract execute

Additional links:

- [The Multiverse upgrade prop is viewable here](https://www.mintscan.io/juno/proposals/28).
- [The v7.0.0 changelog is viewable here](https://github.com/CosmosContracts/juno/releases/tag/v7.0.0). Note however that `v8.0.0` should be used for the upgrade, as it fixes a bug that manifested at the upgrade block.

The target block for this upgrade is [3851750](https://www.mintscan.io/juno/blocks/3851750), which is expected to arrive on _Thursday July 7th at 1700UTC_, +/- 1 hour.

Note that *go 1.18 is now required* - so you will need to remove any older version of go and install the right version if you are building manually.

These are the instructions you will need if you run cosmovisor:

```bash
cd juno
git fetch --tags && git checkout v8.0.0
make build && make install
# this will return commit d0d9f36c5cb4e0128d903bbfd41c96266f0496d8
junod version --long

mkdir -p $DAEMON_HOME/cosmovisor/upgrades/multiverse/bin && cp $HOME/go/bin/junod $DAEMON_HOME/cosmovisor/upgrades/multiverse/bin
# this will return v8.0.0
$DAEMON_HOME/cosmovisor/upgrades/multiverse/bin/junod version
```

Alternatively, you can run the upgrade the old-fashioned way.
