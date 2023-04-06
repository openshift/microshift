#!/usr/bin/env bash

set -euo pipefail
PS4='+ $(date "+%T.%N")\011 '
set -x

SCRIPT_PATH="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")"

RETRIES=5
BACKOFF=20

wait_until() {
    local cmd=$*
    for _ in $(seq "${RETRIES}"); do
        ${cmd} && return 0
        sleep "${BACKOFF}"
    done
    return 1
}

cleanup() {
    oc delete -f "${SCRIPT_PATH}/assets/pod-with-pvc.yaml"
}
trap 'cleanup' EXIT

# TODO: Remove the labels again once https://issues.redhat.com/browse/OCPBUGS-1969 has been fixed upstream
oc label namespaces default "pod-security.kubernetes.io/"{enforce,audit,warn}"-version=v1.24"
oc label namespaces default "pod-security.kubernetes.io/"{enforce,audit,warn}"=privileged"
oc create -f "${SCRIPT_PATH}/assets/pod-with-pvc.yaml"
oc wait --for=condition=Ready --timeout=120s pod/test-pod

set +e
ssh -v "$USHIFT_USER@$USHIFT_IP" "sudo reboot now"
res=$?
set -e

# Allow for `ssh` command errors (255 exit code) like "connection closed by remote host"
# Fail on other errors (coming from the command executed remotely itself)
if [ "${res}" -ne 0 ] && [ "${res}" -ne 255 ]; then
    exit 1
fi

wait_until ssh "$USHIFT_USER@$USHIFT_IP" "true"
# Just check if KAS is up and serving
wait_until oc get node
ssh "$USHIFT_USER@$USHIFT_IP" "sudo /etc/greenboot/check/required.d/40_microshift_running_check.sh"
oc wait --for=condition=Ready --timeout=120s pod/test-pod
