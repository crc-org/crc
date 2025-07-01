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
        ;; 
        *)    # unknown option
        shift 
        ;;
    esac
done

# Prepare resuslts folder
mkdir -p $targetFolder/results

# Run tests
export PATH="$PATH:${HOME}/$targetFolder/bin"
cd $targetFolder/bin
if [ ! -z "$labelFilter" ]
then
    ./integration.test --pull-secret-path="${HOME}/$targetFolder/pull-secret" --bundle-path=$bundleLocation --ginkgo.timeout $suiteTimeout --ginkgo.label-filter "$labelFilter" > integration.results
else
    ./integration.test --pull-secret-path="${HOME}/$targetFolder/pull-secret" --bundle-path=$bundleLocation --ginkgo.timeout $suiteTimeout > integration.results
fi

# Copy results
cd ..
cp bin/integration.results results/integration.results
if [[ -f bin/time-consume.txt ]]; then
  cp bin/time-consume.txt results/time-consume.txt
fi
if which xsltproc &>/dev/null
then
  cat bin/out/integration.xml | xsltproc filter.xsl - > results/$junitFilename
else
  mv bin/out/integration.xml results/$junitFilename
fi
