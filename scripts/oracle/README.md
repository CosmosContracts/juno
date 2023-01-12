# How To Run

```bash
# be in the root directory
# This will reset your ~/.juno directory & start a single instance node
sh ./scripts/test_node.sh clean

# open a new tab / terminal window to install + run the oracle locally
sh ./scripts/oracle/run_local_oracle.sh
# Enter your password at the end, this will be hidden for the keyring

# Wait a little bit for the oracle to submit the data like:
# broadcasting vote exchange_rates=ATOM:10.215474902208682895...

junod q oracle exchange-rate atom --node http://localhost:26657
junod q oracle exchange-rate juno --node http://localhost:26657

# new tab to ensure the oracle querier works between contract<->chain
bash ./scripts/oracle/test_oracle_contract.bash
```
