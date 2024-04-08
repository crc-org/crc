param(
    [Parameter(HelpMessage='When testing a custom bundle we should pass the paht on the target host')]
    $bundleLocation="",
    [Parameter(HelpMessage='To set an specific set of tests based on annotations')]
    $e2eTagExpression="",
    [Parameter(HelpMessage='Name of the folder on the target host under $HOME where all the content will be copied')]
    $targetFolder="crc-e2e",
    [Parameter(HelpMessage='Name for the junit file with the tests results')]
    $junitFilename="e2e-junit.xml",
    [Parameter(HelpMessage='Customize memory for the cluster to run the tests')]
    $crcMemory=""
)

# Prepare run e2e
mv $targetFolder/bin/e2e.test $targetFolder/bin/e2e.test.exe

# Run e2e
$env:PATH="$env:PATH;$env:HOME\$targetFolder\bin;"
$env:SHELL="powershell"
New-Item -ItemType directory -Path "$env:HOME\$targetFolder\results" -Force

# Run tests
$tags="windows"
if ($e2eTagExpression) {
    $tags="$tags && $e2eTagExpression"
}
cd $targetFolder\bin
e2e.test.exe --bundle-location=$bundleLocation --pull-secret-file=$env:HOME\$targetFolder\pull-secret --crc-memory=$crcMemory --cleanup-home=false --godog.tags="$tags" --godog.format=junit > $env:HOME\$targetFolder\results\e2e.results

# Transform results to junit
cd ..
$r = Select-String -Pattern '<?xml version="1.0" encoding="UTF-8"?>' -Path results\e2e.results -list -SimpleMatch | select-object -First 1
Get-Content results\e2e.results | Select -skip ($r.LineNumber -1) > results\$junitFilename
# Copy logs and diagnose
cp -r bin\out\test-results\* results