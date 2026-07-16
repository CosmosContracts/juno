#!/bin/sh

BINARY=${BINARY:-junod}
$BINARY start --minimum-gas-prices 0ujuno --trace
