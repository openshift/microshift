#!/usr/bin/env bash

# reprovision_after_test=false

set -euo pipefail
IFS=$'\n\t'
PS4='+ $(date "+%T.%N")\011 '
set -x

SCRIPT_PATH="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")"

RETRIES=10
BACKOFF=6s

# shellcheck disable=SC2317  # Don't warn about unreachable commands in this function
cleanup() {
    oc delete service microshift-test || true
    oc delete -f "${SCRIPT_PATH}/assets/microshift-test.yaml" || true
    if declare -F firewall::close_port; then firewall::close_port 5678 tcp || true; fi
}
trap 'cleanup' EXIT

declare -F firewall::open_port && firewall::open_port 5678 tcp
oc create -f "${SCRIPT_PATH}/assets/microshift-test.yaml"
oc create service loadbalancer microshift-test --tcp=5678:8080
oc wait pods -l app=microshift-test --for condition=Ready --timeout=60s

for _ in $(seq "${RETRIES}"); do
    set +e
    response=$(curl -i "${USHIFT_IP}":5678 2>&1)
    result=$?
    set -e

    [ ${result} -eq 0 ] &&
        echo "${response}" | grep -q -E "HTTP.*200" &&
        echo "${response}" | grep -q "MicroShift Test" &&
        exit 0

    sleep "${BACKOFF}"
done
exit 1
