#!/bin/bash

set -euo pipefail

# Updates our various build files, CI configs, ... to the latest released go
# version in a given minor release stream (1.x)
# This requires yq and jq in addition to fairly standard shell tools (curl, grep, sed, ...)

golang_base_version=$1
echo "Updating golang version to $golang_base_version"

go mod edit -go ${golang_base_version}
go mod edit -go ${golang_base_version} tools/go.mod
sed -i "s,^\(FROM registry.ci.openshift.org/openshift/release:rhel-8-release-golang-\)1.[0-9]\+,\1${golang_base_version}," images/*/Dockerfile
sed -i "s,^FROM registry.access.redhat.com/ubi8/go-toolset:[.0-9]\+,FROM registry.access.redhat.com/ubi8/go-toolset:${golang_base_version}," images/*/Dockerfile
for f in .github/workflows/*.yml; do
    yq eval --inplace ".jobs.build.strategy.matrix.go[0] = ${golang_base_version}" "$f";
done
