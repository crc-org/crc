#!/bin/bash

# Parameters
bundleLocation=""
e2eTagExpression=""
crcMemory=""
crcBinaryDir="/usr/local/bin"
targetFolder="crc-e2e"
junitFilename="e2e-junit.xml"
while [[ $# -gt 0 ]]; do
    key="$1"
    case $key in
        -bundleLocation)
        bundleLocation="$2"
        shift 
        shift 
        ;;
        -e2eTagExpression)
        e2eTagExpression="$2"
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
        -crcMemory)
        crcMemory="$2"
        shift 
        shift 
        ;;
        -crcBinaryDir)
        crcBinaryDir="$2"
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
tags="linux"
if [ ! -z "$e2eTagExpression" ]
then
    tags="$tags && $e2eTagExpression"
fi
cd $targetFolder/bin
./e2e.test --bundle-location=$bundleLocation --pull-secret-file="${HOME}/$targetFolder/pull-secret" --crc-binary=$crcBinaryDir --cleanup-home=false --crc-memory=$crcMemory --godog.tags="$tags" --godog.format=junit > "${HOME}/$targetFolder/results/e2e.results"

# Transform results to junit
cd ..
init_line=$(grep -n '<?xml version="1.0" encoding="UTF-8"?>' results/e2e.results | awk '{split($0,n,":"); print n[1]}')
if which xsltproc &>/dev/null
then
  tail -n +$init_line results/e2e.results | xsltproc filter.xsl - > results/$junitFilename
else
  tail -n +$init_line results/e2e.results > results/$junitFilename
fi
# Copy logs and diagnose
cp -r bin/out/test-results/* results
