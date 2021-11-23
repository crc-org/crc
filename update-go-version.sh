#!/bin/bash

set -euo pipefail

# Updates our various build files, CI configs, ... to the latest released go
# version in a given minor release stream (1.x)
# This requires yq and jq in addition to fairly standard shell tools (curl, grep, sed, ...)

golang_base_version=$1
latest_version=$(curl --location --silent  'https://go.dev/dl/?mode=json&include=all' | jq -r '.[].files[].version'  |uniq | sed -e 's/go//' |sort -V |grep ${golang_base_version}|tail -1)
echo "Updating golang version to $latest_version"

go mod edit -go ${golang_base_version}
sed -i "s,^FROM registry.ci.openshift.org/openshift/release:golang-1\... AS builder\$,FROM registry.ci.openshift.org/openshift/release:golang-${golang_base_version} AS builder," images/openshift-ci/Dockerfile
sed -i "s,^FROM registry.access.redhat.com/ubi8/go-toolset:[.0-9]\+ as builder\$,FROM registry.access.redhat.com/ubi8/go-toolset:${latest_version} as builder," images/build-e2e/Dockerfile
sed -i "s,^FROM registry.access.redhat.com/ubi8/go-toolset:[.0-9]\+ as builder\$,FROM registry.access.redhat.com/ubi8/go-toolset:${latest_version} as builder," images/build-integration/Dockerfile
sed -i "s/GOVERSION: .*\$/GOVERSION: \"${latest_version}\"/" .circleci/config.yml
sed -i "s/^GO_VERSION=.*$/GO_VERSION=${latest_version}/" centos_ci.sh
appveyor_go_version=$(echo $golang_base_version | tr -d .)
sed -i 's/set PATH="C:\\go[0-9]\+"/set PATH="C:\\go'${appveyor_go_version}'"/' ./appveyor.yml
