#!/bin/bash
set -euxo pipefail

BASEDIR=$(dirname "$0")
OUTPUT=$1
CODESIGN_IDENTITY=${CODESIGN_IDENTITY:-mock}
PRODUCTSIGN_IDENTITY=${PRODUCTSIGN_IDENTITY:-mock}
NO_CODESIGN=${NO_CODESIGN:-0}

function sign() {
  if [ "${NO_CODESIGN}" -eq "1" ]; then
    return
  fi
  local opts=""
  entitlements="${BASEDIR}/$(basename "$1").entitlements"
  if [ -f "${entitlements}" ]; then
      opts="--entitlements ${entitlements}"
  fi
  codesign --deep --sign "${CODESIGN_IDENTITY}" --options runtime --force ${opts} "$1"
}

binDir="${BASEDIR}/root/Applications/CodeReady Containers.app/Contents/Resources"

version=$(cat "${BASEDIR}/VERSION")

sign "${binDir}/crc"
sign "${binDir}/admin-helper-darwin"
sign "${binDir}/crc-driver-hyperkit"

sign "${BASEDIR}/root/Applications/CodeReady Containers.app"

codesign --verify --verbose "${binDir}/hyperkit"

pkgbuild --identifier com.redhat.crc --version ${version} \
  --scripts "${BASEDIR}/darwin/scripts" \
  --root "${BASEDIR}/root" \
  --install-location / \
  --component-plist "${BASEDIR}/components.plist" \
  "${OUTPUT}/crc.pkg"

productbuild --distribution "${BASEDIR}/darwin/Distribution" \
  --resources "${BASEDIR}/darwin/Resources" \
  --package-path "${OUTPUT}" \
  "${OUTPUT}/crc-unsigned.pkg"
rm "${OUTPUT}/crc.pkg"

if [ ! "${NO_CODESIGN}" -eq "1" ]; then
  productsign --sign "${PRODUCTSIGN_IDENTITY}" "${OUTPUT}/crc-unsigned.pkg" "${OUTPUT}/crc-macos-amd64.pkg"
else
  mv "${OUTPUT}/crc-unsigned.pkg" "${OUTPUT}/crc-macos-amd64.pkg"
fi
