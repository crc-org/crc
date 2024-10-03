param(
    [Parameter(HelpMessage='When testing a custom bundle we should pass the paht on the target host')]
    $bundleLocation="",
    [Parameter(HelpMessage='Name of the folder on the target host under $HOME where all the content will be copied')]
    $targetFolder="crc-integration",
    [Parameter(HelpMessage='Name for the junit file with the tests results')]
    $junitFilename="integration-junit.xml",
    [Parameter(HelpMessage='Test suite fails if it does not complete within the specified timeout. Default 90m')]
    $suiteTimeout="90m",
    [Parameter(HelpMessage='Filter tests to be executed based on label expression')]
    $labelFilter=""
)

# Prepare run e2e
mv $targetFolder/bin/integration.test $targetFolder/bin/integration.test.exe

# Run e2e
$env:PATH="$env:PATH;$env:HOME\$targetFolder\bin;"
$env:SHELL="powershell"
$targetFolderDir = "$env:HOME\$targetFolder"
$resultsDir = "$targetFolderDir\results"

New-Item -ItemType directory -Path "$resultsDir" -Force

# Run tests
cd $targetFolder\bin

if ($labelFilter) {
    integration.test.exe --pull-secret-path="$targetFolderDir\pull-secret" --bundle-path=$bundleLocation --ginkgo.timeout $suiteTimeout --ginkgo.label-filter "$labelFilter" > integration.results
} else {
    integration.test.exe --pull-secret-path="$targetFolderDir\pull-secret" --bundle-path=$bundleLocation --ginkgo.timeout $suiteTimeout > integration.results
}


# Copy results
cd ..
cp bin\integration.results results\integration.results
$prejunit = "$resultsDir\$junitFilename.pre"
cp bin\out\integration.xml $prejunit

$xslt = New-Object System.Xml.Xsl.XslCompiledTransform;
$xslt.load("$targetFolderDir\filter.xsl")
$xslt.transform( "$prejunit", "$resultsDir\$junitFilename" )
rm "$prejunit"
