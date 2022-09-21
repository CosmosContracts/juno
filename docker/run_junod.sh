#!/bin/sh

if test -n "$1"; then
    # need -R not -r to copy hidden files
    cp -R "$1/.juno" /root
fi

mkdir -p /root/log
junod start --grpc.address tcp://0.0.0.0:9090 --rpc.laddr tcp://0.0.0.0:26657 --trace