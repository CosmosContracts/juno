# Scripts

This section contains scripts that are useful for altering the Juno project during development. They should be run from the root of the Juno directory unless specified otherwise.

---

## Setup

To setup a Juno node, run the following command:

```sh
make install
make init
```

This will install the dependencies and initialize the node with default values. You can override the default values by setting the environment variables specified in the `scripts/init.sh` file.

## Protobuffers

To generate code files from their respective `.proto` files, run the following command:

```sh
make proto-all
```

which runs the following code generators:

```sh
./scripts/buf/buf-gogo.sh
./scripts/buf/buf-pulsar.sh
./scripts/buf/buf-openapi.sh
```

You can also run the scripts manually. In some cases, you have to use the `sudo` command for the recursive copy (from temp to the proto file locations in `x/` directory).

---

## Local Testing Environment

To startup 1 or multiple Juno instances quickly, you will use the `docker-compose.yml` file.
By default it is configured to run 1 instance of Juno on your local changes. You can override the default values by setting another node in the services section of the `docker-compose.yml` file.

Start/stop/remove the Juno instances:

```sh
docker-compose up -d
```

```sh
docker-compose stop
```

```sh
docker-compose down
```

## Statesync

Statesync is a quick way to sync a Juno node without having to download an entire snapshot (>100GB in most cases). Our snapshot uses the `pebbledb` database and it is automatically configured for you on run.

```bash
bash ./scripts/statesync.bash
```
