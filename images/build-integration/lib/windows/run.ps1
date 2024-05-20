param(
    [Parameter(HelpMessage='When testing a custom bundle we should pass the paht on the target host')]
    $bundleLocation="",
    [Parameter(HelpMessage='Name of the folder on the target host under $HOME where all the content will be copied')]
    $targetFolder="crc-integration",
    [Parameter(HelpMessage='Name for the junit file with the tests results')]
    $junitFilename="integration-junit.xml",
    [Parameter(HelpMessage='Test suite fails if it does not complete within the specified timeout. Default 90m')]
    $suiteTimeout="90m"
)

# Prepare run e2e
mv $targetFolder/bin/integration.test $targetFolder/bin/integration.test.exe

# Run e2e
$env:PATH="$env:PATH;$env:HOME\$targetFolder\bin;"
$env:SHELL="powershell"
New-Item -ItemType directory -Path "$env:HOME\$targetFolder\results" -Force

# Run tests
cd $targetFolder\bin
if ($bundleLocation) {
    $env:BUNDLE_PATH="$bundleLocation"
}
# We need to copy the pull-secret to the target folder
$env:PULL_SECRET_PATH="$env:HOME\$targetFolder\pull-secret"
integration.test.exe --ginkgo.timeout $suiteTimeout > integration.results

# Copy results
cd ..
cp bin\integration.results results\integration.results
cp bin\out\integration.xml results\$junitFilename