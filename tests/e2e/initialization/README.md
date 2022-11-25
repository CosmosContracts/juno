# initializaion package

## Motivation

This package contains all logic necessary for initializing configuration
data either for a new chain or a single node via Docker containers.

The motivation for doing this via Docker is to be able to initialize
configs of any Juno version.

For example, while the latest Juno version is v11,
we might want to spin up a chain of v10 and test the upgrade.

Additionally, there are known file permission errors when initializing
configurations as non-root. This is troublesome both in CI and locally.
Doing this via Docker instead, allows us to initialize these files as
a root user, bypassing the file permission issues.

## Structure

Each folder in `tests/e2e/initialization` corresponds to a standalone script.
At the time of this writing, we have the following scripts/folders:
    - `chain` - for initializing a full chain
    - `node` - for initializing a single node

All initialization scripts share a common `init.Dockerfile` that
takes an argument `E2E_SCRIPT_NAME`. By providing the desired script
name to the Dockerfile, we are able to build the image that can run
any of these local scripts

## Scripts

### Initializing a Chain (`chain`)

From root folder:

```sh
make docker-build-e2e-init-chain
```

This script will build a Docker image that runs a script in the `chain` package
and initializes all configuration files necessary for starting up an e2e chain.

#### Running The Container

When running a container with the specified script, it must mount a folder on a volume
to have all configuration files produced.

Additionally, it takes the following arguments:

- `--data-dir`
  - the location of where the configuration data is written inside
    the container (string)
- `--chain-id`
  - the id of the chain (string)
- `--config`
  - serialized node configurats (e.g. Pruning and Snapshot options).
    These correspond to the stuct `NodeConfig`, located in
    `tests/e2e/chain/config.go` The number of initialized
    validators on the new chain corresponds to the number of
    `NodeConfig`s provided by this parameter
- `--voting-period`
  - The configurable voting period duration for the chain

```go
    tmpDir, _ := os.MkdirTemp("", "juno-e2e-testnet-")

 initResource, _ = s.dkrPool.RunWithOptions(
  &dockertest.RunOptions{
   Name:       fmt.Sprintf("%s", chainId),
   Repository: s.dockerImages.InitRepository,
   Tag:        s.dockerImages.InitTag,
   NetworkID:  s.dkrNet.Network.ID,
   Cmd: []string{
    fmt.Sprintf("--data-dir=%s", tmpDir),
    fmt.Sprintf("--chain-id=%s", chainId),
    fmt.Sprintf("--config=%s", nodeConfigBytes),
    fmt.Sprintf("--voting-period=%v", votingPeriodDuration),
   },
   User: "root:root",
   Mounts: []string{
    fmt.Sprintf("%s:%s", tmpDir, tmpDir), 
   },
  },
  noRestart,
 )
```

#### Container Output

Assumming that a the container was correctly mounted on a volume,
it produces the following:

- `juno-test-< chain id >-encode` file
  - This is encoded metadata about the newly created chain with its nodes
- `juno-test-< chain id >` folder
  - For every `NodeCondig` provided to the container, it will produce a folder
    with the respective node configs

Example:

```sh
$:/tmp/juno-e2e-testnet-1167397304 $ ls
juno-test-a  juno-test-a-encode

$:/tmp/juno-e2e-testnet-1167397304/juno-test-a $ cd  juno-test-a

$:/tmp/juno-e2e-testnet-1167397304/juno-test-a $ cd  juno-test-a-juno-00

$:/tmp/juno-e2e-testnet-1167397304/juno-test-a/juno-test-a-juno-00 $ ls
config  data  keyring-test  wasm
```

- Here we mounted the container on
`/tmp/juno-e2e-testnet-1167397304/juno-test`as a volume
- < chain id > = "a"
- 4 `NodeConfig`s were provided via the `--config` flag
- `juno-test-a-encode` output file corresponds to the serialized `Chain` struct
defined in `tests/e2e/chain/chain.go`

### Initializing a Node (`node`)

```sh
make docker-build-e2e-init-node
```

This script will build a Docker image that runs a script in the `node` package
and initializes all data necessary for starting up a new node.