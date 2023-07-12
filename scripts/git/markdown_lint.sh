#!/usr/bin/env bash

if ! command -v markdownlint &> /dev/null ; then
    echo "markdownlint not installed" >&2
    echo "please check https://www.npmjs.com/package/markdownlint" >&2
    exit 1
fi

res="$(markdownlint . --disable MD013 MD010 2>&1 echo)"
word_count=`echo $res | wc -w`

if [ $word_count -gt 0 ]; then
  echo "markdownlint failed"
  echo -e "$res"
  exit 2
fi