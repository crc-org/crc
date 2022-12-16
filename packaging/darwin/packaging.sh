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
  codesign --deep --sign "${CODESIGN_IDENTITY}" --options runtime --timestamp --force ${opts} "$1"
}

function signAppBundle() {
  if [ "${NO_CODESIGN}" -eq "1" ]; then
    return
  fi
  entitlements=$(sed -e 's| |_|g' <<< "${BASEDIR}/$(basename "$1").entitlements")
  if [ ! -f "${entitlements}" ]; then
    echo "ERROR: need entitlement file: ${entitlements}"
    return
  fi

  frameworks=$(find "$1"/Contents/Frameworks -depth -type d -name "*.framework" -or -name "*.dylib" -or -type f -perm +111)
  echo "${frameworks}" | xargs -t -I % codesign --deep --sign "${CODESIGN_IDENTITY}" --options runtime --timestamp % || true

  # sign the .app bundles inside $1/Contents/Frameworks
  frameworks=$(find "$1"/Contents/Frameworks -depth -type d -name "*.app" -perm +111)
  echo "${frameworks}" | xargs -t -I % codesign --deep --sign "${CODESIGN_IDENTITY}" --options runtime --timestamp --force % || true

  # finally sign $1 app bundle
  codesign --deep --sign "${CODESIGN_IDENTITY}" --options runtime --timestamp --force --entitlements "${entitlements}" "$1"
}

binDir="${BASEDIR}/root/Applications/Red Hat OpenShift Local.app/Contents/Resources"

version=$(cat "${BASEDIR}/VERSION")

pkgbuild --analyze --root ${BASEDIR}/root ${BASEDIR}/components.plist
plutil -replace BundleIsRelocatable -bool NO ${BASEDIR}/components.plist

sign "${binDir}/crc"
sign "${binDir}/crc-admin-helper-darwin"
sign "${binDir}/vfkit"

signAppBundle "${BASEDIR}/root/Applications/Red Hat OpenShift Local.app"

sudo chmod +sx "${binDir}/crc-admin-helper-darwin"

pkgbuild --identifier com.redhat.crc --version ${version} \
  --scripts "${BASEDIR}/scripts" \
  --root "${BASEDIR}/root" \
  --install-location / \
  --component-plist "${BASEDIR}/components.plist" \
  "${OUTPUT}/crc.pkg"

productbuild --distribution "${BASEDIR}/Distribution" \
  --resources "${BASEDIR}/Resources" \
  --package-path "${OUTPUT}" \
  "${OUTPUT}/crc-unsigned.pkg"
rm "${OUTPUT}/crc.pkg"

if [ ! "${NO_CODESIGN}" -eq "1" ]; then
  productsign --sign "${PRODUCTSIGN_IDENTITY}" "${OUTPUT}/crc-unsigned.pkg" "${OUTPUT}/crc-macos-installer.pkg"
else
  mv "${OUTPUT}/crc-unsigned.pkg" "${OUTPUT}/crc-macos-installer.pkg"
fi
