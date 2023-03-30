#!/usr/bin/env bash

set -euo pipefail
IFS=$'\n\t'
set -x

SCRIPT_PATH="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")"

# shellcheck disable=SC2317  # Don't warn about unreachable commands in this function
cleanup() {
    oc delete route hello-microshift
    oc delete service hello-microshift
    oc delete -f "${SCRIPT_PATH}/assets/hello-microshift.yaml"
    firewall::close_port tcp:80
}
trap cleanup EXIT

firewall::open_port tcp:80
oc create -f "${SCRIPT_PATH}/assets/hello-microshift.yaml"
oc expose pod hello-microshift
oc expose svc hello-microshift --hostname hello-microshift.cluster.local
oc wait pods -l app=hello-microshift --for condition=Ready --timeout=60s

retries=3
backoff=3s
for _ in $(seq 1 "${retries}"); do
    RESPONSE=$(curl -i http://hello-microshift.cluster.local --resolve "hello-microshift.cluster.local:80:${IP}" 2>&1)
    RESULT=$?
    [ $RESULT -eq 0 ] &&
        echo "${RESPONSE}" | grep -q -E "HTTP.*200" &&
        echo "${RESPONSE}" | grep -q "Hello MicroShift" &&
        exit 0

    sleep "${backoff}"
done

exit 1
