# CW20 ICS20 IBC Transfer

Transfer a CW20 token from 1 chain to another over the ICS20 standard transfer port.

## Steps

Install cw20-ics20 and cw20 contracts
or download from the release page pre-compiled

<https://github.com/CosmWasm/cw-plus/releases>

```bash
git clone git@github.com:CosmWasm/cw-plus.git
cd cw-plus/
sh scripts/optimizer.sh # installs contracts cw20_base & cw20_ics20
```

## Running

run 2 nodes from the Juno repo + relayer

```bash
# Start both chains
CHAIN_ID="local-1" HOME_DIR="~/.juno1/" TIMEOUT_COMMIT="500ms" CLEAN=true sh scripts/test_node.sh
CHAIN_ID="local-2" HOME_DIR="~/.juno2/" CLEAN=true RPC=36657 REST=2317 PROFF=6061 P2P=36656 GRPC=8090 GRPC_WEB=8091 ROSETTA=8081 TIMEOUT_COMMIT="500ms" sh scripts/test_node.sh

# then run the run script to upload, init, and execute on the above chains
sh ./scripts/hermes/cw20/run.sh

# relay from the CW20_ICS contract
A_PORT=wasm.juno1nc5tatafv6eyq7llkr2gv50ff9e22mnf70qgjlv737ktmt4eswrq68ev2p CHANNEL_VERSION=ics20-1 sh ./scripts/hermes/start.sh
# press 'y' if prompted to use default invocation
```
