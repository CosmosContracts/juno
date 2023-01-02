# Junø - Mainnet

![JUNOVERSE 34](https://user-images.githubusercontent.com/79812965/134063436-6f1bda5c-56f3-4bf3-a3a0-2b93f24217b1.png)

## Syncing a node

For a complete description of how to sync a node, including all upgrades, check out the [Juno Documentation site](https://docs.junonetwork.io/validators/mainnet-upgrades).

In the `juno-1` folder you can find a number of upgrade files in all caps.

Those that begin `0`, e.g. `0100_MONETA_UPGRADE.md` refer to the pre-attack `juno-1`. 

Those that begin `1`, e.g. `1100_v3_1_0_UPGRADE.md` refer to the post-attack, relaunched `juno-1`.

Those that begin `2`, e.g. `2100_vx_x_x_UPGRADE.md` refer to the post-July 28 attack, relaunched `juno-1`.

## Original launch docs

**Note these are left for historical reasons and are no longer relevant to running the chain.**

_Planned Start Time: October 1st at 15:00 UTC._

**Genesis File**

[Genesis File](/juno-1/genesis.json):

```bash
   curl -s  https://raw.githubusercontent.com/CosmosContracts/mainnet/main/juno-1/genesis.json >~/.juno/config/genesis.json
```

**Genesis sha256**

```bash
sha256sum ~/.juno/config/genesis.json
a5c08e53aca0390c45def85a6d16c0e7176bd0026b0a465aff5d1896ec0134a1
```

**junod version**

```bash
$ junod version --long
name: juno
server_name: junod
version: HEAD-e507450f2e20aa4017e046bd24a7d8f1d3ca437a
commit: e507450f2e20aa4017e046bd24a7d8f1d3ca437a
```

**Seed node**

[Full seed nodes list](/juno-1/seeds.txt).

```
2484353dab0b2c1275765b8ffa2c50b3b36158ca@seed-node.junochain.com:26656
```

## Setup

**Prerequisites:** Make sure to have [Golang >=1.17](https://golang.org/).

#### Build from source

You need to ensure your gopath configuration is correct. If the following **'make'** step does not work then you might have to add these lines to your .profile or .zshrc in the users home folder:

```bash
export GOROOT=/usr/local/go
export GOPATH=$HOME/go
export GO111MODULE=on
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
```

```bash
git clone https://github.com/CosmosContracts/juno
cd juno
git checkout v1.0.0
make build && make install
```

This will build and install `junod` binary into `$GOBIN`.

Note: When building from source, it is important to have your `$GOPATH` set correctly. When in doubt, the following should do:

```bash
mkdir ~/go
export GOPATH=~/go
```

### Minimum hardware requirements

- 32-64GB RAM
- 500GB of disk space
- 1.4 GHz amd64 CPU

## Setup validator node

Below are the instructions to generate & submit your genesis transaction

### Generate genesis transaction (pre-launch only)

Similar to Osmosis, only nodes that received the airdrop will be able to validate. Others will be able to join the validator set at a later date.

1. Initialize the Juno directories and create the local genesis file with the correct
   chain-id

   ```bash
   junod init <MONIKER-NAME> --chain-id=juno-1
   ```

2. Create a local key pair (you should use the same key associated with you airdropped account)

   ```bash
   junod keys add <key-name>
   ```

   Note: if you're using an offline key for signing (for example, with a Ledger), do this with `junod keys add <KEY-NAME> --pubkey <YOUR-PUBKEY>`. For the rest of the transactions, you will use the `--generate-only` flag and sign them offline with `junod tx sign`.

3. Download the pre-genesis file:

   ```bash
   curl -s  https://raw.githubusercontent.com/CosmosContracts/mainnet/main/juno-1/pre-genesis.json >~/.juno/config/genesis.json
   ```

   Find your account in the `juno-1/pre-genesis.json` file. The balance of your airdrop is what you'll be able to use with your validator.

4. Create the gentx, replace `<KEY-NAME>` and `<BALANCE>`:

   ```bash
   junod gentx <KEY-NAME> <BALANCE>ujuno --chain-id=juno-1
   ```

   If all goes well, you will see a message similar to the following:

   ```bash
   Genesis transaction written to "/home/user/.juno/config/gentx/gentx-******.json"
   ```

### Submit genesis transaction

- Fork this repo into your Github account

- Clone your repo using:

  ```bash
  git clone https://github.com/<YOUR-GITHUB-USERNAME>/mainnet
  ```

- Copy the generated gentx json file to `<REPO-PATH>/juno-1/gentx/`

  ```bash
  cd mainnet
  cp ~/.juno/config/gentx/gentx*.json ./juno-1/gentx/
  ```

- Commit and push to your repo
- Create a PR onto https://github.com/CosmosContracts/mainnet
- Only PRs from individuals / groups with a history of successfully running validator nodes and that have initial juno balance from the stakedrop will be accepted. This is to ensure the network successfully starts on time.

#### Running in production

**Note, we'll be going through some upgrades soon after Juno mainnet. Consider using [Cosmovisor](https://docs.junochain.com/validators/setting-up-cosmovisor) to make your life easier.**

Download Genesis file when the time is right. Put it in your `/home/<YOUR-USERNAME>/.juno` folder.

Create a systemd file for your Juno service:

```bash
sudo nano /etc/systemd/system/junod.service
```

Copy and paste the following and update `<YOUR-USERNAME>`:

```bash
Description=Juno daemon
After=network-online.target

[Service]
User=<YOUR_USERNAME>
ExecStart=/home/<YOUR-USERNAME>/go/bin/junod start --home /home/<YOUR-USERNAME>/.juno
Restart=on-failure
RestartSec=3
LimitNOFILE=4096

[Install]
WantedBy=multi-user.target
```

**This assumes `$HOME/.juno` to be your directory for config and data. Your actual directory locations may vary.**

Enable and start the new service:

```bash
sudo systemctl enable junod
sudo systemctl start junod
```

Check status:

```bash
junod status
```

Check logs:

```bash
journalctl -u junod -f
```

### Learn more

- [Juno Documentation](https://docs.junochain.com)
- [Juno Community Discord](https://discord.gg/QcWPfK4gJ2)
- [Juno Community Telegram](https://t.me/joinchat/R7QKD0ltosphNWNk)
