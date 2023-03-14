# Scripts

This section contains scripts that are useful for altering the Juno project during development. They should be run from the root of the Juno directory unless specified otherwise.

---

## Statesync

Statesync is a quick way to sync a Juno node without having to download an entire snapshot (>100GB in most cases). Our snapshot uses the `pebbledb` database and it is automatically configured for you on run.

```bash
bash ./scripts/statesync.bash
```

## Proto files generation

To generate protobuf files from their respective `.proto` files, run the following command:

```bash
sh ./scripts/protocgen.sh
```

You can also run it manually if you open the file. In some cases, you have to use the `sudo` command for the recursive copy (from temp to the proto file locations in `x/`).

---

## Fast Local Testing Environment

To startup 1 or 2 Juno instances quickly, you will use the `test_node.sh` script like so:

```bash
CHAIN_ID="local-1" HOME_DIR="~/.juno1/" TIMEOUT_COMMIT="500ms" CLEAN=true sh scripts/test_node.sh

CHAIN_ID="local-2" HOME_DIR="~/.juno2/" CLEAN=true RPC=36657 REST=2317 PROFF=6061 P2P=36656 GRPC=8090 GRPC_WEB=8091 ROSETTA=8081 TIMEOUT_COMMIT="500ms" sh scripts/test_node.sh
```

It does not require Docker. If you wish to run only 1 instance, the top line is the default for standard port mappings. Using the variable CLEAN fresh installs the tip of the branch and also resets the database and all config files for the home directory.

## Local Relaying

We provide a simple relaying script to transfer packets between 2 local chains. This will auto setup, connect, and create the channel from local-1 to local-2 test nodes.

Here is how to use it:

```bash
# Start both chains
CHAIN_ID="local-1" HOME_DIR="~/.juno1/" TIMEOUT_COMMIT="500ms" CLEAN=true sh scripts/test_node.sh
CHAIN_ID="local-2" HOME_DIR="~/.juno2/" CLEAN=true RPC=36657 REST=2317 PROFF=6061 P2P=36656 GRPC=8090 GRPC_WEB=8091 ROSETTA=8081 TIMEOUT_COMMIT="500ms" sh scripts/test_node.sh

# start the relayer
sh ./scripts/hermes/start.sh

```
