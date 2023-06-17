#!/usr/bin/env bash

set -euo pipefail
PS4='+ $(date "+%T")\011 '
set -x

SCRIPT_PATH="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")"
ASSETS="${SCRIPT_PATH}/assets"

TEST_STRING="FOOBAR"

SC="${ASSETS}/storage-class-thin.yaml"
SNAPSHOT="${ASSETS}/snapshot.yaml"
PRODUCER="${ASSETS}/pod-with-pvc-thin.yaml"
CONSUMER="${ASSETS}/pod-with-pvc-thin-consumer.yaml"
POD_NAME="test-pod-thin"

cleanup(){
    oc delete --ignore-not-found=true -f "${SC}" -f "${SNAPSHOT}" -f "${PRODUCER}" -f "${CONSUMER}"
}
trap cleanup EXIT

oc create -f "${SC}"

oc create -f "${PRODUCER}"
oc wait --for=condition=Ready pod/"${POD_NAME}" --timeout=30s

oc exec -it pod/"${POD_NAME}" -- sh -c 'echo '"${TEST_STRING}"' > /vol/file.txt'

oc delete pod "${POD_NAME}"
oc wait --for=delete "pod/${POD_NAME}" --timeout=30s

oc create -f "${SNAPSHOT}"
oc wait --for=jsonpath='{.status.readyToUse}'=true volumesnapshot my-snap --timeout 30s

oc delete pvc test-claim-thin
oc wait --for=delete pvc/test-claim-thin --timeout=30s

oc create -f "${CONSUMER}"
oc wait --for=condition=Ready "pod/${POD_NAME}" --timeout=30s

GOT_STRING="$(oc exec -it "pod/${POD_NAME}" -- sh -c 'cat /vol/file.txt' | tr -d '[:space:]')"
if [ "${TEST_STRING}" != "${GOT_STRING}" ]; then
    >&2 echo "SNAPSHOT SMOKE-TEST: FAIL"
    exit 1
fi
echo 'SNAPSHOT SMOKE-TEST: SUCCESS'
