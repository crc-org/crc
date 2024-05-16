# Chocolatey package for CRC (non GUI)

## Pre-requisite

The host machine needs to have [Chocolatey](https://chocolatey.org/) installed and should be able to build `crc`. Follow [chocolatey install guide](https://chocolatey.org/install) to install it.

## Steps to build the package

Build steps for the chocolatey package is incorporated in the `Makefile` under the `choco` target.

1. Run `make choco`:
```
PS> make choco
rm -rf /cygdrive/c/Users/anath/redhat/crc/docs/build
rm -f packaging/darwin/Distribution
rm -f packaging/darwin/Resources/welcome.html
rm -f packaging/darwin/scripts/postinstall
rm -rf packaging/darwin/root/
rm -rf packaging/windows/msi
rm -f out/windows-amd64/split
rm -f packaging/rpm/crc.spec images/rpmbuild/Containerfile
rm -rf out
rm -f C:\Users\anath\go/bin/crc
rm -rf release
GOARCH=amd64 GOOS=windows go build -tags "containers_image_openpgp" -ldflags="-X github.com/crc-org/crc/v2/pkg/crc/version.crcVersion=2.12.0 -X github.com/crc-org/crc/v2/pkg/crc/version.ocpVersion=4.11.18 -X github.com/crc-org/crc/v2/pkg/crc/version.okdVersion=4.11.0-0.okd-2022-11-05-030711 -X github.com/crc-org/crc/v2/pkg/crc/version.commitSha=40b730ce " -o out/windows-amd64/crc.exe  ./cmd/crc
go build --tags="build" -ldflags="-X github.com/crc-org/crc/v2/pkg/crc/version.crcVersion=2.12.0 -X github.com/crc-org/crc/v2/pkg/crc/version.ocpVersion=4.11.18 -X github.com/crc-org/crc/v2/pkg/crc/version.okdVersion=4.11.0-0.okd-2022-11-05-030711 -X github.com/crc-org/crc/v2/pkg/crc/version.commitSha=40b730ce " -o out/windows-amd64/crc-embedder  ./cmd/crc-embedder
out/windows-amd64/crc-embedder download --goos=windows packaging/chocolatey/crc/tools
0 B / 97.26 MiB [______________________________________________________________________________________________________________________________________________________________________________________] 0.00% ? p/s
cp out/windows-amd64/crc.exe packaging/chocolatey/crc/tools/crc.exe
cd packaging/chocolatey/crc && choco pack
Chocolatey v1.2.0
Attempting to build package from 'crc.nuspec'.
Successfully created package 'C:\Users\anath\redhat\crc\packaging\chocolatey\crc\crc.2.12.0.nupkg'
```

2. Above command will generated `crc.nupkg` in the `packaging/chocolatey/crc` folder

## Steps to push the package to chocolatey community repo

1. Create [account](https://community.chocolatey.org/account/Register) in the chocolatey community feed.
3. Copy the API Key from the [accounts page](https://community.chocolatey.org/account) (Click on the Show API Key button).
4. Associate your API key with community feed:
```
PS> choco apikey --key <account_api_key> --source https://push.chocolatey.org/
```
5. Push the package using:
```
PS> choco push crc.2.11.0.nupkg --source https://push.chocolatey.org/
```

> Note: Chocolatey has a moderation process and there might be a need to answer queries from the moderators, keep an eye on the [package queue](https://community.chocolatey.org/packages).
