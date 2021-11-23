FROM ghcr.io/cosmoscontracts/junod:latest

RUN sh /scripts/setup_junod.sh

# rest server
EXPOSE 1317
# tendermint p2p
EXPOSE 26656
# tendermint rpc
EXPOSE 26657

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT [ "/entrypoint.sh" ]

CMD [ "start", "--rpc.laddr", "tcp://0.0.0.0:26657", "--trace" ]