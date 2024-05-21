#!/bin/bash

# Parameters
bundleLocation=""
targetFolder="crc-integration"
junitFilename="integration-junit.xml"
suiteTimeout="90m"
labelFilter=""
while [[ $# -gt 0 ]]; do
    key="$1"
    case $key in
        -bundleLocation)
        bundleLocation="$2"
        shift 
        shift 
        ;;
        -targetFolder)
        targetFolder="$2"
        shift 
        shift 
        ;;
        -junitFilename)
        junitFilename="$2"
        shift 
        shift 
        ;;
        -suiteTimeout)
        suiteTimeout="$2"
        shift 
        shift 
        ;;
        -labelFilter)
        labelFilter="$2"
        shift 
        shift 
        *)    # unknown option
        shift 
        ;;
    esac
done

# Prepare resuslts folder
mkdir -p $targetFolder/results

# Run tests
export PATH="$PATH:${HOME}/$targetFolder/bin"
export PULL_SECRET_PATH="${HOME}/$targetFolder/pull-secret"
if [ ! -z "$bundleLocation" ]
then
    export BUNDLE_PATH="$bundleLocation"
fi
cd $targetFolder/bin
if [ ! -z "$labelFilter" ]
then
    ./integration.test --ginkgo.timeout $suiteTimeout --ginkgo.label-filter $labelFilter > integration.results
else
    ./integration.test --ginkgo.timeout $suiteTimeout > integration.results
fi

# Copy results
cd ..
cp bin/integration.results results/integration.results
cp bin/out/integration.xml results/$junitFilename