#!/usr/bin/env bash

set -euo pipefail
IFS=$'\n\t'
set -x

SCRIPT_PATH="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")"

# shellcheck disable=SC2317  # Don't warn about unreachable commands in this function
cleanup() {
    oc delete route hello-microshift || true
    oc delete service hello-microshift || true
    oc delete -f "${SCRIPT_PATH}/assets/hello-microshift.yaml" || true
    firewall::close_port 80 tcp || true
}
trap cleanup EXIT

firewall::open_port 80 tcp
oc create -f "${SCRIPT_PATH}/assets/hello-microshift.yaml"
oc expose pod hello-microshift
oc expose svc hello-microshift --hostname hello-microshift.cluster.local
oc wait pods -l app=hello-microshift --for condition=Ready --timeout=60s

retries=3
backoff=3s
for _ in $(seq 1 "${retries}"); do
    RESPONSE=$(curl -i http://hello-microshift.cluster.local --resolve "hello-microshift.cluster.local:80:${USHIFT_IP}" 2>&1)
    RESULT=$?
    [ $RESULT -eq 0 ] &&
        echo "${RESPONSE}" | grep -q -E "HTTP.*200" &&
        echo "${RESPONSE}" | grep -q "Hello MicroShift" &&
        exit 0

    sleep "${backoff}"
done

exit 1
