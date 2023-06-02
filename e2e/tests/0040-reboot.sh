#!/usr/bin/env bash

# reprovision_after_test=false

set -euo pipefail
PS4='+ $(date "+%T.%N")\011 '
set -x

SCRIPT_PATH="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")"

RETRIES=60
INTERVAL=5
SUCCESSFUL_TRIES=3

get_remote_boot_timestamp() {
    ssh "${USHIFT_USER}@${USHIFT_IP}" "date -d \"\$(uptime -s)\" +%s"
}

wait_for_system_up() {
    local prev_boot_start="${1}"
    local successes=0

    for _ in $(seq "${RETRIES}"); do
        curr_boot_start=$(get_remote_boot_timestamp || true)

        if [[ "${curr_boot_start}" -gt "${prev_boot_start}" ]]; then
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

prev_boot_start=$(get_remote_boot_timestamp)
set +e
ssh -v "${USHIFT_USER}@${USHIFT_IP}" "sudo reboot now"
res=$?
set -e

# Allow for `ssh` command errors (255 exit code) like "connection closed by remote host"
# Fail on other errors (coming from the command executed remotely itself)
if [ "${res}" -ne 0 ] && [ "${res}" -ne 255 ]; then
    exit 1
fi

wait_for_system_up "${prev_boot_start}"
ssh "${USHIFT_USER}@${USHIFT_IP}" "sudo /etc/greenboot/check/required.d/40_microshift_running_check.sh"
oc wait --for=condition=Ready --timeout=120s pod/test-pod
