#!/usr/bin/env bash

if ! command -v golangci-lint &> /dev/null ; then
    echo "golangci-lint not installed or available in the PATH" >&2
    echo "please check https://github.com/golangci/golangci-lint" >&2
    exit 1
fi

res="$(golangci-lint run ./... --allow-parallel-runners --concurrency 1 --fix)"
word_count=`echo $res | wc -w`

if [ $word_count -gt 1 ]; then
  echo "golangci-lint failed"
  echo -e "$res"

  exit 2
fi