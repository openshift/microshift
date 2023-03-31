#!/usr/bin/env bash

set -euo pipefail
IFS=$'\n\t'
set -x

SCRIPT_PATH="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")"

# shellcheck disable=SC2317  # Don't warn about unreachable commands in this function
cleanup() {
    oc delete service hello-microshift || true
    oc delete -f "${SCRIPT_PATH}/assets/hello-microshift.yaml" || true
    firewall::close_port 5678 tcp || true
}
trap cleanup EXIT

firewall::open_port 5678 tcp
oc create -f "${SCRIPT_PATH}/assets/hello-microshift.yaml"
oc create service loadbalancer hello-microshift --tcp=5678:8080
oc wait pods -l app=hello-microshift --for condition=Ready --timeout=60s

retries=3
backoff=3s
for _ in $(seq 1 "${retries}"); do
    RESPONSE=$(curl -i "${USHIFT_IP}":5678 2>&1)
    RESULT=$?
    [ $RESULT -eq 0 ] &&
        echo "${RESPONSE}" | grep -q -E "HTTP.*200" &&
        echo "${RESPONSE}" | grep -q "Hello MicroShift" &&
        exit 0

    sleep "${backoff}"
done

exit 1
