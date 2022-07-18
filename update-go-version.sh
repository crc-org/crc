#!/bin/bash

set -euo pipefail

# Updates our various build files, CI configs, ... to the latest released go
# version in a given minor release stream (1.x)
# This requires yq and jq in addition to fairly standard shell tools (curl, grep, sed, ...)

golang_base_version=$1
latest_version=$(curl --location --silent  'https://go.dev/dl/?mode=json&include=all' | jq -r '.[].files[].version'  |uniq | sed -e 's/go//' |sort -V |grep ${golang_base_version}|tail -1)
echo "Updating golang version to $latest_version"

go mod edit -go ${golang_base_version}
go mod edit -go ${golang_base_version} tools/go.mod
sed -i "s,^FROM registry.ci.openshift.org/openshift/release:golang-1\... AS builder\$,FROM registry.ci.openshift.org/openshift/release:golang-${golang_base_version} AS builder," images/openshift-ci/Dockerfile
sed -i "s,^FROM registry.access.redhat.com/ubi8/go-toolset:[.0-9]\+ as builder\$,FROM registry.access.redhat.com/ubi8/go-toolset:${latest_version} as builder," images/build-e2e/Dockerfile
sed -i "s,^FROM registry.access.redhat.com/ubi8/go-toolset:[.0-9]\+ as builder\$,FROM registry.access.redhat.com/ubi8/go-toolset:${latest_version} as builder," images/build-integration/Dockerfile
for f in .github/workflows/*.yml; do
    yq eval --inplace ".jobs.build.strategy.matrix.go[0] = ${golang_base_version}" "$f";
done
