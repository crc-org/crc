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

git clone "${REPO_ROOT_DIR}" "${TMP_DIFFROOT}"

make -C "${TMP_DIFFROOT}" update-go-version

echo "Diffing ${REPO_ROOT_DIR} against tree with freshly updated golang version"
ret=0
git --no-pager -C ${TMP_DIFFROOT} diff --exit-code 2>&1 >/dev/null || ret=1
#diff -Naupr "${REPO_ROOT_DIR}/vendor" "${TMP_DIFFROOT}/vendor" || ret=1

if [[ $ret -eq 0 ]]
then
  echo "${REPO_ROOT_DIR} up to date."
else
  echo "${REPO_ROOT_DIR} is out of date. Please run make update-go-version"
  exit 1
fi
