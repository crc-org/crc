#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

export GO111MODULE=on

if [[ -n $(git status -s vendor/ tools/vendor/) ]]; then
        echo 'vendor/ directory has uncommitted changes, please check `git status vendor`'
        exit 1
fi

make vendor

go mod verify
cd tools && go mod verify && cd ..

echo "Diffing $(pwd)"

if git diff --exit-code vendor go.mod go.sum tools/vendor tools/go.mod tools/go.sum; then
  echo "$(pwd) up to date."
else
  echo "$(pwd) is out of date. Please run make vendor"
  exit 1
fi
