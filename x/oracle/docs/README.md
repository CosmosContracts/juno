# Oracle Setup Documentation

## PreReqs

### Install Dependencies

```sh
sudo apt update && sudo apt upgrade -y
sudo apt-get install make gcc jq git
```

### Install Go

```sh
curl -fsSL https://golang.org/dl/go1.19.3.linux-amd64.tar.gz | sudo tar -xzC /usr/local
```

### Prepare Go Paths

```sh
echo "export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin" >> $HOME/.bashrc

source ~/.bashrc

go version  # should output "go version go1.19.3 linux/amd64"
```

### Build Price Feeder

```sh
cd ~
git clone https://github.com/CosmosContracts/juno.git
cd juno/pricefeeder
make install

# Copy config to your location for future use & editing
cp config.toml $HOME/.juno/oracle-config.toml

price-feeder version # sdk: v0.45.11 and go1.19.*
```

---

## Auto-Generate script

Edit variables `in the script` & run the script for the sections you need

```sh
#!/bin/sh

VALOPER_ADDR=""
FEEDER_ADDR=""
CHAIN_ID="" # juno-1 or uni-5
KEYRING_PASS=""

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
cp config.toml $HOME/.juno/oracle-config.toml
sed -i \
's/0.0001stake/0.025ujuno/g;
 s/address = "juno1w20tfhnehc33rgtm9tg8gdtea0svn7twfnyqee"/address = "'"$FEEDER_ADDR"'"/g;
 s/validator = "junovaloper1w20tfhnehc33rgtm9tg8gdtea0svn7twkwj0zq"/validator = "'"$VALOPER_ADDR"'"/g;
 s/chain_id = "test-1"/chain_id = "'"$CHAIN_ID"'"/g;
 s/"chain_id", "test-1"/"chain_id", "'"$CHAIN_ID"'"/g; \
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
WantedBy=multi-user.target" | sudo tee -a /etc/systemd/system/oracle.Service
"

sudo systemctl daemon-reload
sudo systemctl enable oracle
sudo systemctl start oracle
```

```sh
sudo journalctl -fu oracle
```

---

## Screen

```sh
sudo apt update && sudo apt upgrade -y
sudo apt-get screen
```

## Set Variables

```sh
FEEDER_ADDR=juno_address # edit to your value
```

```sh
VALOPER_ADDR=junovaloper_address # edit to your value
```

```sh
CHAIN_ID=juno-1 # edit to correct chain
```

## Edit Configuration File

```sh
sed -i \
's/0.0001stake/0.025ujuno/g;
 s/address = "juno1w20tfhnehc33rgtm9tg8gdtea0svn7twfnyqee"/address = "'"$FEEDER_ADDR"'"/g;
 s/validator = "junovaloper1w20tfhnehc33rgtm9tg8gdtea0svn7twkwj0zq"/validator = "'"$VALOPER_ADDR"'"/g;
 s/chain_id = "test-1"/chain_id = "'"$CHAIN_ID"'"/g;
 s/"chain_id", "test-1"/"chain_id", "'"$CHAIN_ID"'"/g;
 s/websocket = "stream.binance.com:9443"/websocket = "fstream.binance.com:443"/g' \
~/.juno/oracle-config.toml
```

### Attach to new screen

```sh
screen -S price-feeder
```

```sh
price-feeder ~/.juno/oracle-config.toml --log-level debug
```

### Detach from the price-feeder screen

```sh
ctrl-a + d: Detach a screen session without stopping it.
screen -r price-feeder: Reattach a detached screen session.
```

---

## Systemd

```sh
sudo apt update && sudo apt upgrade -y
sudo apt install git -y
```

## Edit Variables

```sh
sudo nano juno/oracle/scripts/systemd.sh
```

```sh
VALOPER_ADDR="junovaloper_address_here" #valoper address here
FEEDER_ADDR="junofeeder_address_here" #feeder wallet address here
CHAIN_ID="juno-1" #chain_id here
KEYRING_PASS="mykeyringpass" #keyring password here, anything if using test keyring
```

### Run Systemd Script

```sh
cd ~
./juno/oracle/scripts/systemd.sh
```

### Check Status

```sh
sudo journalctl -fu oracle
```

---

## Docker

```sh
sudo apt update -y && apt upgrade -y && apt autoremove -y
sudo apt install docker.io docker-compose -y
```

### Clone Respository

```sh
git clone https://github.com/CosmosContracts/juno
cd juno/price-feeder
```

### Build Image

```sh
docker build -t price-feeder .
```

### Run Container

```sh
docker run -d --name price-feeder price-feeder
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
