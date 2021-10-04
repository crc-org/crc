#!/bin/sh

# Vars
BINARY=e2e.test
if [[ ${PLATFORM} == 'windows' ]]; then
    BINARY=e2e.test.exe
fi
BINARY_PATH="/opt/crc/bin/${PLATFORM}-amd64/${BINARY}"

# Running options
CLEANUP_HOME="${CLEANUP_HOME:-"true"}"
# e2e tag picking
E2E_TAG_EXPRESSION="${E2E_TAG_EXPRESSION:-""}"
# Review this when go 1.16 with embed support
FEATURES_PATH=/opt/crc/features
TESTDATA_PATH=/opt/crc/testdata
UX_RESOURCES_PATH=/opt/crc/ux

# Results
RESULTS_PATH="${RESULTS_PATH:-/output}"
# results file name
RESULTS_FILE="${RESULTS_FILE:-"e2e"}"

if [ "${DEBUG:-}" = "true" ]; then
    set -xuo 
fi

# Validate conf
validate=true
[[ -z "${TARGET_HOST+x}" ]] \
    && echo "TARGET_HOST required" \
    && validate=false

[[ -z "${TARGET_HOST_USERNAME+x}" ]] \
    && echo "TARGET_HOST_USERNAME required" \
    && validate=false

[[ -z "${TARGET_HOST_KEY_PATH+x}" && -z "${TARGET_HOST_PASSWORD+x}" ]] \
    && echo "TARGET_HOST_KEY_PATH or TARGET_HOST_PASSWORD required" \
    && validate=false

[[ -z "${PULL_SECRET_FILE_PATH+x}" ]] \
    && echo "PULL_SECRET_FILE_PATH required" \
    && validate=false

[[ -z "${BUNDLE_VERSION+x}" && -z "${BUNDLE_LOCATION+x}" ]] \
    && echo "BUNDLE_VERSION or BUNDLE_LOCATION required" \
    && validate=false

[[ $validate == false ]] && exit 1

# Define remote connection
REMOTE="${TARGET_HOST_USERNAME}@${TARGET_HOST}"
if [[ ! -z "${TARGET_HOST_DOMAIN+x}" ]]; then
    REMOTE="${TARGET_HOST_USERNAME}@${TARGET_HOST_DOMAIN}@${TARGET_HOST}"
fi

# Set SCP / SSH command with pass or key
NO_STRICT='-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null'
if [[ ! -z "${TARGET_HOST_KEY_PATH+x}" ]]; then
    SCP="scp -r ${NO_STRICT} -i ${TARGET_HOST_KEY_PATH}"
    SSH="ssh ${NO_STRICT} -i ${TARGET_HOST_KEY_PATH}"
else
    SCP="sshpass -p ${TARGET_HOST_PASSWORD} scp -r ${NO_STRICT}" \
    SSH="sshpass -p ${TARGET_HOST_PASSWORD} ssh ${NO_STRICT}"
fi

echo "Copy resources to target"
# Create execution folder 
EXECUTION_FOLDER="/Users/${TARGET_HOST_USERNAME}/crc-e2e"
if [[ ${PLATFORM} == 'linux' ]]; then
    EXECUTION_FOLDER="/home/${TARGET_HOST_USERNAME}/crc-e2e"
fi
DATA_FOLDER="${EXECUTION_FOLDER}/out"
if [[ ${PLATFORM} == 'windows' ]]; then
    # Todo change for powershell cmdlet
    $SSH "${REMOTE}" "powershell.exe -c New-Item -ItemType directory -Path ${EXECUTION_FOLDER}/bin"
else
    $SSH "${REMOTE}" "mkdir -p ${EXECUTION_FOLDER}/bin"
fi

# Copy crc-e2e binary and pull-secret
# Review this when go 1.16 with embed support
$SCP "${BINARY_PATH}" "${REMOTE}:${EXECUTION_FOLDER}/bin"
$SCP "${PULL_SECRET_FILE_PATH}" "${REMOTE}:${EXECUTION_FOLDER}/pull-secret"
$SCP "${FEATURES_PATH}" "${REMOTE}:${EXECUTION_FOLDER}/bin"
# Applescripts
REMOTE_RESOURCES_PATH=/workspace/test/extended/crc
if [[ ${PLATFORM} == 'windows' ]]; then
    # Todo change for powershell cmdlet
    $SSH "${REMOTE}" "powershell.exe -c New-Item -ItemType directory -Path ${REMOTE_RESOURCES_PATH}"
    $SCP "${UX_RESOURCES_PATH}" "${REMOTE}:${REMOTE_RESOURCES_PATH}"

else
    # This relies on:  cat /etc/synthetic.conf 
    #workspace  Users/crcqe/crc-e2e/workspace as symbolic link
    $SSH "${REMOTE}" "mkdir -p ${EXECUTION_FOLDER}/workspace/test/extended/crc"
    $SCP "${UX_RESOURCES_PATH}" "${REMOTE}:${EXECUTION_FOLDER}/workspace/test/extended/crc"
fi
# Testdata files
$SCP "${TESTDATA_PATH}" "${REMOTE}:${EXECUTION_FOLDER}"

echo "Running e2e tests"
# e2e envs
if [[ ${PLATFORM} == 'windows' ]]; then
    # BINARY_EXEC="(New-Object -ComObject "Shell.Application").minimizeall(); \$env:SHELL=\"powershell\"; "
    BINARY_EXEC="\$env:SHELL=\"powershell\"; "
fi
# e2e running options
if [[ ! -z "${BUNDLE_LOCATION+x}" ]]; then
    OPTIONS="--bundle-location=${BUNDLE_LOCATION} "
else
    OPTIONS="--bundle-location=\"\" --bundle-version=${BUNDLE_VERSION} "
fi
if [[ ${PLATFORM} == 'windows' ]]; then
    OPTIONS+="--pull-secret-file=C:\\Users\\${TARGET_HOST_USERNAME}\\crc-e2e\\pull-secret "
else
    OPTIONS+="--pull-secret-file=${EXECUTION_FOLDER}/pull-secret "
fi
if [[ ${PLATFORM} == 'macos' ]]; then
    PLATFORM="darwin"
fi
if [ "${CLEANUP_HOME}" = "false" ]; then
    OPTIONS+="--cleanup-home=false "
fi
if [ -n "${INSTALLER_PATH}" ]; then
    OPTIONS+="--installer-path=${INSTALLER_PATH} "
fi
if [ -n "${USER_PASSWORD}" ]; then
    OPTIONS+="--user-password=${USER_PASSWORD} "
fi
OPTIONS+="--godog.tags=\"@${PLATFORM} && ${E2E_TAG_EXPRESSION}\" --godog.format=junit "
BINARY_EXEC+="cd ${EXECUTION_FOLDER}/bin && ./${BINARY} ${OPTIONS} > ${RESULTS_FILE}.results"
# Execute command remote
$SSH ${REMOTE} ${BINARY_EXEC}

echo "Getting e2e tests results and logs"
# Get results
mkdir -p "${RESULTS_PATH}"
$SCP "${REMOTE}:${EXECUTION_FOLDER}/bin/${RESULTS_FILE}.results" "${RESULTS_PATH}"
$SCP "${REMOTE}:${EXECUTION_FOLDER}/bin/out/test-results" "${RESULTS_PATH}"
# xunit cleanup on results
pushd "${RESULTS_PATH}"
init_line=$(grep -n '<?xml version="1.0" encoding="UTF-8"?>' ${RESULTS_FILE}.results | awk '{split($0,n,":"); print n[1]}')
tail -n +$init_line ${RESULTS_FILE}.results > ${RESULTS_FILE}.xml
popd

echo "Cleanup target"
# Cleanup
# Review this when go 1.16 with embed support
if [[ ${PLATFORM} == 'windows' ]]; then
    # Todo change for powershell cmdlet
    $SSH "${REMOTE}" "rm -r /workspace"
fi
$SSH "${REMOTE}" "rm -r ${EXECUTION_FOLDER}"