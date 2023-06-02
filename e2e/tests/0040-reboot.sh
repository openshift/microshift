#!/usr/bin/env bash

# reprovision_after_test=false

set -euo pipefail
PS4='+ $(date "+%T.%N")\011 '
set -x

SCRIPT_PATH="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")"

RETRIES=25
INTERVAL=10
SUCCESSFUL_TRIES=3

wait_until() {
    successes=0
    local cmd=$*
    for _ in $(seq "${RETRIES}"); do
        if ${cmd}; then
          successes=$((successes+=1))
        else 
          successes=0
        fi
        if [[ "${successes}" -ge "${SUCCESSFUL_TRIES}" ]]; then
            return 0
        fi
        sleep "${INTERVAL}"
    done
    return 1
}

cleanup() {
    oc delete -f "${SCRIPT_PATH}/assets/pod-with-pvc.yaml"
}
trap 'cleanup' EXIT

oc create -f "${SCRIPT_PATH}/assets/pod-with-pvc.yaml"
oc wait --for=condition=Ready --timeout=120s pod/test-pod

set +e
ssh -v "${USHIFT_USER}@${USHIFT_IP}" "sudo reboot now"
res=$?
set -e

# Allow for `ssh` command errors (255 exit code) like "connection closed by remote host"
# Fail on other errors (coming from the command executed remotely itself)
if [ "${res}" -ne 0 ] && [ "${res}" -ne 255 ]; then
    exit 1
fi

wait_until ssh "${USHIFT_USER}@${USHIFT_IP}" "true"
ssh "${USHIFT_USER}@${USHIFT_IP}" "sudo /etc/greenboot/check/required.d/40_microshift_running_check.sh"
oc wait --for=condition=Ready --timeout=120s pod/test-pod
