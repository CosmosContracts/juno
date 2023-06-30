#!/usr/bin/env bash

function runTidy() {
  go mod tidy -v
  if [ $? -ne 0 ]; then
    echo "go mod tidy failed"
    exit 2
  fi

  git diff --exit-code go.* &> /dev/null
  if [ $? -ne 0 ]; then
      echo "root go.mod or go.sum differs, please re-add it to your commit"
      exit 3
  fi
}

runTidy

cd interchaintest/
runTidy
