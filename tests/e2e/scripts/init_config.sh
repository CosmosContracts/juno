#!/bin/sh
sed -i '114,117d' config.toml
address=`sed '214!d' /tmp/juno-e2e-testnet/juno-test-a/juno-test-a-node-prune-default-snapshot-state-sync-from/config/genesis.json`
validator=`sed '215!d' /tmp/juno-e2e-testnet/juno-test-a/juno-test-a-node-prune-default-snapshot-state-sync-from/config/genesis.json`
var1=${address#*juno}
var2=${var1%\"*}
var3=${validator#*juno}
var4=${var3%\"*}
echo "address = \"juno$var2\"" >> config.toml
echo "chain_id = \"juno-test-a\"" >> config.toml
echo "validator = \"juno$var4\"" >> config.toml
echo "prefix = \"juno\"" >> config.toml