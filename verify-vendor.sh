#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

export GO111MODULE=on

readonly REPO_ROOT_DIR="$(git rev-parse --show-toplevel 2> /dev/null)"
readonly TMP_DIFFROOT="$(mktemp -d "${REPO_ROOT_DIR}"/tmpdiffroot.XXXXXX)"

cleanup() {
  rm -rf "${TMP_DIFFROOT}"
}

trap "cleanup" EXIT SIGINT

cleanup

if [[ -n $(git status -s vendor/) ]]; then
        echo 'vendor/ directory has uncommitted changes, please check `git status vendor`'
        exit 1
fi

mkdir "${TMP_DIFFROOT}"
cp -aR "${REPO_ROOT_DIR}/vendor" "${TMP_DIFFROOT}"

make vendor

echo "Diffing ${REPO_ROOT_DIR} against freshly generated codegen"
ret=0
diff -Naupr "${REPO_ROOT_DIR}/vendor" "${TMP_DIFFROOT}/vendor" || ret=1

# Restore working tree state
rm -fr "${REPO_ROOT_DIR}/vendor"
cp -aR "${TMP_DIFFROOT}"/* "${REPO_ROOT_DIR}"

if [[ $ret -eq 0 ]]
then
  echo "${REPO_ROOT_DIR} up to date."
else
  echo "${REPO_ROOT_DIR} is out of date. Please run make vendor"
  exit 1
fi
