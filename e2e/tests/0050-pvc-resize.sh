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

oc create -f "${SCRIPT_PATH}/assets/pod-with-pvc.yaml"
oc wait --for=condition=Ready --timeout=120s pod/test-pod

RESIZE_TO=2Gi
TIME_OUT=3m
oc patch pvc test-claim -p '{"spec":{"resources":{"requests":{"storage":"'${RESIZE_TO}'"}}}}'
oc wait --timeout ${TIME_OUT} --for=jsonpath="{.spec.resources.requests.storage}"=${RESIZE_TO} pvc/test-claim
