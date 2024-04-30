#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

export GO111MODULE=on

if [[ -n $(git status -s vendor/) ]]; then
        echo 'vendor/ directory has uncommitted changes, please check `git status vendor`'
        exit 1
fi

make vendor

go mod verify

echo "Diffing $(pwd)"
git diff --exit-code vendor go.mod go.sum

if [[ $? -eq 0 ]]
then
  echo "$(pwd) up to date."
else
  echo "$(pwd) is out of date. Please run make vendor/make toolsvendor"
  exit 1
fi
