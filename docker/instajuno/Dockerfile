FROM faddat/archlinux as builder

ENV GOPATH /go
ENV PATH $PATH:/go/bin


RUN pacman -Syyu --noconfirm base-devel go gcc git wget rust && \
        git clone https://github.com/cosmoscontracts/juno && \
        cd juno && \
        make

RUN wget http://165.232.134.27:8080/lucina-snapshot.tar.gz && \
        tar xvf lucina-snapshot.tar.gz

RUN wget https://github.com/CosmosContracts/testnets/raw/main/lucina/genesis.json

RUN git clone https://github.com/CosmWasm/wasmvm && \
        cd wasmvm && \
        make build

FROM faddat/archlinux

COPY --from=builder /go/bin/junod /usr/bin/junod 
COPY --from=builder /data /data
COPY --from=builder /genesis.json /genesis.json
COPY --from=builder /wasmvm/api/libwasmvm.so /lib/libwasmvm.so

# With volume mounts and the like one can avoid this madness once your node is up, and it can be suitable for runing a validator. 
# For a validator: just don't run the CMD (docker run)
CMD junod init instajuno$RANDOM && cp /genesis.json ~/.juno/config && cp -r /data/* ~/.juno/data && junod start --p2p.persistent_peers 3a46840ad6ef9f014fb8157409eef6808806c861@74.207.254.130:26656,ed292d66c239f3b86f19cbb1d7c6c125464b87d0@66.42.91.67:26656,ff68dc9f1808d9d497a769719cecd2785e093776@207.244.245.6:26656,6ec2b1b9f0c279875de1b493278e1c54d2de98f5@104.236.76.128:26656,74667752912bb51325492682ab97bbff8edb7101@165.232.134.27:26656  --p2p.seed_mode true

