# Oracle Setup Documentation

## Install Base Dependencies

```sh
# Ubuntu
sudo apt update && sudo apt upgrade -y
sudo apt-get install make gcc jq git
```

### Install Go 1.19

```sh
curl -fsSL https://golang.org/dl/go1.19.3.linux-amd64.tar.gz | sudo tar -xzC /usr/local

# `go version` should output "go version go1.19.3 linux/amd64"
```

### Prepare Go Paths

```sh
# Only run this if your go path is not yet setup.
# Check this with `echo $PATH`

echo "export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin" >> $HOME/.bashrc

# Reload the bashrc file to the current terminal
source $HOME/.bashrc
```

### Build Price Feeder

```sh
cd ~
git clone https://github.com/CosmosContracts/juno.git
cd ./juno/price-feeder
make install

# Copy config to your location for future use & editing
cp config.example.toml $HOME/.juno/oracle-config.toml

# If you get `cp: cannot create regular file '$HOME/.juno/oracle-config.toml': No such file or directory`, 
# run: `junod init [moniker]`. Then run the cp command again

price-feeder version # sdk: v0.45.11 and go1.19.*
# If it is not found, your go path is not set
```

---

## Systemd

```sh
sudo apt update && sudo apt upgrade -y
sudo apt install git -y
```

## Edit Variables

```sh
nano $HOME/juno/scripts/oracle/systemd.sh
```

```sh
#!/bin/sh

VALOPER_ADDR="junovaloper1..."
FEEDER_ADDR="juno1..."
CHAIN_ID="juno-1"
# anything if you use the test keyring in your oracle-config.toml
KEYRING_PASS="mykeyringpass"

sudo apt update && sudo apt upgrade -y
sudo apt-get install make gcc jq 

# Install GO
GO_VER="1.19.2"
cd $HOME
wget "https://golang.org/dl/go$GO_VER.linux-amd64.tar.gz"
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf "go$GO_VER.linux-amd64.tar.gz"
rm "go$GO_VER.linux-amd64.tar.gz"
echo "export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin" >> $HOME/.profile
source $HOME/.profile

# Install price feeder
cd pricefeeder
make install

# Setup price feeder config
cp $HOME/juno/price-feeder/config.example.toml $HOME/.juno/oracle-config.toml

sed -i \
's/0.0001stake/0.025ujuno/g;
 s/address = "juno1w20tfhnehc33rgtm9tg8gdtea0svn7twfnyqee"/address = "'"$FEEDER_ADDR"'"/g;
 s/validator = "junovaloper1w20tfhnehc33rgtm9tg8gdtea0svn7twkwj0zq"/validator = "'"$VALOPER_ADDR"'"/g;
 s/chain_id = "test-1"/chain_id = "'"$CHAIN_ID"'"/g;
 s/"chain_id", "test-1"/"chain_id", "'"$CHAIN_ID"'"/g' \
$HOME/.juno/oracle-config.toml


# Setup systemd service
echo "[Unit]
Description=juno-price-feeder
After=network.target

[Service]
Type=simple
User=$USER
ExecStart=$HOME/go/bin/price-feeder $HOME/.juno/oracle-config.toml --log-level debug
Restart=on-abort
LimitNOFILE=65535
Environment=\"PRICE_FEEDER_PASS=$KEYRING_PASS\"

[Install]
WantedBy=multi-user.target" | sudo tee /etc/systemd/system/oracle.service

sudo systemctl daemon-reload
sudo systemctl enable oracle
sudo systemctl start oracle
```

### Run System Script

```sh
sh $HOME/juno/scripts/oracle/systemd.sh
```

### Check the values are correct

```sh
cat /etc/systemd/system/oracle.service
```

### Check Status

```sh
sudo journalctl -fu oracle --output cat
```

### Stop the Systemd service

```sh
sudo systemctl stop oracle
```

---

## Docker

```sh
sudo apt update -y && apt upgrade -y && apt autoremove -y
sudo apt install containerd docker.io docker-compose -y

sudo systemctl start docker 
```

### Clone Respository

```sh
git clone https://github.com/CosmosContracts/juno
cd $HOME/juno
```

### Build Image

```sh
docker build -f ./price-feeder/price-feeder.Dockerfile -t price-feeder .
```

### Run Container

```sh
# This requires use of the test keyring as we mount to the keyring-test volume.
# We also mount as read only to the oracle config from above to /oracle-config.toml in the container
# network host allows us to query the host machine RPC (https://localhost:26657)
docker run -d --name price-feeder \
    -e PRICE_FEEDER_PASS="mypass" \
    -v /root/.juno/keyring-test:/root/.juno/keyring-test:ro \
    --mount type=bind,source=/root/.juno/oracle-config.toml,target=/oracle-config.toml,readonly \
    --network="host" \
    price-feeder /oracle-config.toml
```

### Check Logs

```sh
docker logs price-feeder
```

### Shell Into Container

```sh
docker exec -it price-feeder /bin/sh
```

### Stop or Remove Container

```sh
docker stop price-feeder
docker rm price-feeder
```
