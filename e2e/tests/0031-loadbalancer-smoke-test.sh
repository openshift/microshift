#!/usr/bin/env bash

set -euo pipefail
IFS=$'\n\t'
PS4='+ $(date "+%T.%N")\011 '
set -x

SCRIPT_PATH="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")"

# shellcheck disable=SC2317  # Don't warn about unreachable commands in this function
cleanup() {
    oc delete service hello-microshift || true
    oc delete -f "${SCRIPT_PATH}/assets/hello-microshift.yaml" || true
    if declare -F firewall::close_port; then firewall::close_port 5678 tcp || true; fi
}
RESULT=1
# explicit exit after cleanup, to not have cleanup override RESULT
trap 'cleanup; exit $RESULT' EXIT

declare -F firewall::open_port && firewall::open_port 5678 tcp
oc create -f "${SCRIPT_PATH}/assets/hello-microshift.yaml"
oc create service loadbalancer hello-microshift --tcp=5678:8080
oc wait pods -l app=hello-microshift --for condition=Ready --timeout=60s

retries=3
backoff=3s
for _ in $(seq 1 "${retries}"); do
    RESPONSE=$(curl -i "${USHIFT_IP}":5678 2>&1)
    RESULT=$?
    if [ $RESULT -eq 0 ] &&
        echo "${RESPONSE}" | grep -q -E "HTTP.*200" &&
        echo "${RESPONSE}" | grep -q "Hello MicroShift"; then
        RESULT=0
        break
    fi
    sleep "${backoff}"
done
