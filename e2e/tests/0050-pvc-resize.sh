#!/usr/bin/env bash

set -euo pipefail
PS4='+ $(date "+%T.%N")\011 '
set -x

SCRIPT_PATH="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")"
POD_MANIFEST="${SCRIPT_PATH}/assets/kustomizations/patches/pvc-thick"

RETRIES=5
RETRY_WAIT=20

wait_until() {
    local cmd=$*
    for _ in $(seq "${RETRIES}"); do
        ${cmd} && return 0
        sleep "${RETRY_WAIT}"
    done
    return 1
}

MANIFEST="$(mktemp -d)/manifest.yaml"
oc kustomize "${POD_MANIFEST}" | tee "${MANIFEST}"

CLAIM_NAME="$(yq 'select(.kind == "PersistentVolumeClaim") | .metadata.name' "${MANIFEST}")"
POD_NAME="$(yq 'select(.kind == "Pod") | .metadata.name' "${MANIFEST}")"

cleanup() {
    oc delete -f "${MANIFEST}"
}
trap 'cleanup' EXIT

oc apply -f "${MANIFEST}"
oc wait --for=condition=Ready --timeout=120s pod/"${POD_NAME}"

RESIZE_TO=2Gi
TIME_OUT=3m
oc patch pvc "${CLAIM_NAME}" -p '{"spec":{"resources":{"requests":{"storage":"'${RESIZE_TO}'"}}}}'
oc wait --timeout ${TIME_OUT} --for=jsonpath="{.spec.resources.requests.storage}"=${RESIZE_TO} pvc/"${CLAIM_NAME}"
