#!/bin/sh

# Vars
BINARY=integration.test
if [[ ${PLATFORM} == 'windows' ]]; then
    BINARY=integration.test.exe
fi
BINARY_PATH="/opt/crc/bin/${PLATFORM}-amd64/${BINARY}"

# Results
RESULTS_PATH="${RESULTS_PATH:-/output}"


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
EXECUTION_FOLDER="/Users/${TARGET_HOST_USERNAME}/crc-integration"
if [[ ${PLATFORM} == 'linux' ]]; then
    EXECUTION_FOLDER="/home/${TARGET_HOST_USERNAME}/crc-integration"
fi
DATA_FOLDER="${EXECUTION_FOLDER}/out"
if [[ ${PLATFORM} == 'windows' ]]; then
    # Todo change for powershell cmdlet
    $SSH "${REMOTE}" "powershell.exe -c New-Item -ItemType directory -Path ${EXECUTION_FOLDER}/bin"
else
    $SSH "${REMOTE}" "mkdir -p ${EXECUTION_FOLDER}/bin"
fi

# Copy crc-integration binary and pull-secret
$SCP "${BINARY_PATH}" "${REMOTE}:${EXECUTION_FOLDER}/bin"
$SCP "${PULL_SECRET_FILE_PATH}" "${REMOTE}:${EXECUTION_FOLDER}/pull-secret"

echo "Running integration tests"
# Run integration cmd
if [[ ! -z "${BUNDLE_LOCATION+x}" ]] && [[ -n "${BUNDLE_LOCATION}" ]]; then
    BUNDLE_PATH="${BUNDLE_LOCATION}"
else
    BUNDLE_PATH="embedded"
fi
# Review when pwsh added as powershell supported
if [[ ${PLATFORM} == 'windows' ]]; then
    BINARY_EXEC="cd ${EXECUTION_FOLDER}/bin; "
    BINARY_EXEC+="\$env:SHELL='powershell'; "
    BINARY_EXEC+="\$env:PULL_SECRET_PATH='${EXECUTION_FOLDER}/pull-secret'; "
    BINARY_EXEC+="\$env:BUNDLE_PATH='${BUNDLE_PATH}'; "
else 
    BINARY_EXEC="cd ${EXECUTION_FOLDER}/bin && "
    BINARY_EXEC+="PULL_SECRET_PATH=${EXECUTION_FOLDER}/pull-secret "
    BINARY_EXEC+="BUNDLE_PATH=${BUNDLE_PATH} "
fi
BINARY_EXEC+="./${BINARY} > integration.results"
# Execute command remote
$SSH "${REMOTE}" "${BINARY_EXEC}"

echo "Getting integration results and logs"
# Get results
mkdir -p "${RESULTS_PATH}"
$SCP "${REMOTE}:${EXECUTION_FOLDER}/bin/integration.results" "${RESULTS_PATH}"
$SCP "${REMOTE}:${EXECUTION_FOLDER}/bin/out/integration.xml" "${RESULTS_PATH}"

echo "Cleanup target"
# Clenaup
$SSH "${REMOTE}" "rm -r ${EXECUTION_FOLDER}"