#!/usr/bin/env bash

# reprovision_after_test=false

set -euo pipefail
IFS=$'\n\t'
PS4='+ $(date "+%T.%N")\011 '
set -x

SCRIPT_PATH="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")"

RETRIES=3
BACKOFF=3s

# shellcheck disable=SC2317  # Don't warn about unreachable commands in this function
cleanup() {
    oc delete route hello-microshift || true
    oc delete service hello-microshift || true
    oc delete -f "${SCRIPT_PATH}/assets/hello-microshift.yaml" || true
    if declare -F firewall::close_port; then firewall::close_port 80 tcp || true; fi
}
trap 'cleanup' EXIT

declare -F firewall::open_port && firewall::open_port 80 tcp

oc create -f "${SCRIPT_PATH}/assets/hello-microshift.yaml"
oc expose pod hello-microshift
oc expose svc hello-microshift --hostname hello-microshift.cluster.local
oc wait pods -l app=hello-microshift --for condition=Ready --timeout=60s

for _ in $(seq "${RETRIES}"); do
    set +e
    response=$(curl -i http://hello-microshift.cluster.local --resolve "hello-microshift.cluster.local:80:${USHIFT_IP}" 2>&1)
    result=$?
    set -e

    [ ${result} -eq 0 ] &&
        echo "${response}" | grep -q -E "HTTP.*200" &&
        echo "${response}" | grep -q "Microshift Test" &&
        exit 0

    sleep "${BACKOFF}"
done
exit 1
