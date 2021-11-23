#!/bin/sh

IFS=","
for addr in "${JUNOD_SEED_ADDR}"; do
    if [ -z $addr ] ; then
        continue
    fi
    junod add-genesis-account "$addr" "1000000000$STAKE,1000000000$FEE"
done

exec /usr/bin/junod $@
