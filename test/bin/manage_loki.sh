#!/bin/bash
#
# This script should be run on the hypervisor to set up an ephemeral
# Loki server for those tests that require metrics checking.

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

DEFAULT_HOST_PORT="3100"

usage() {
    cat - <<EOF
${BASH_SOURCE[0]} (start|stop) [port]

  -h           Show this help.

start [port]: Start Loki.
             Uses port ${DEFAULT_HOST_PORT} on the host by default.
             The container name will be loki-<host_port>.

stop [port]: Stop Loki.
            Uses port ${DEFAULT_HOST_PORT} by default to identify the container.
            The container name is assumed to be loki-<host_port>.

EOF
}

action_stop() {
    local host_port="${1:-${DEFAULT_HOST_PORT}}"
    local container_name="loki-${host_port}"
    echo "Stopping Loki container ${container_name}"
    podman stop "${container_name}" > /dev/null || true
    podman rm --force "${container_name}" > /dev/null || true
}

action_start() {
    local host_port="${1:-${DEFAULT_HOST_PORT}}"
    local container_name="loki-${host_port}"

    echo "Stopping previous instance of Loki container ${container_name} (if any)"
    action_stop "${host_port}"

    echo "Starting Loki container ${container_name} on host port ${host_port}"
    podman run -d --rm --name "${container_name}" \
        -p "${host_port}:3100" \
        docker.io/grafana/loki > /dev/null
}

if [ $# -eq 0 ]; then
    usage
    exit 1
fi
action="${1}"
shift

case "${action}" in
    start|stop)
        "action_${action}" "$@"
        ;;
    -h)
        usage
        exit 0
        ;;
    *)
        usage
        exit 1
        ;;
esac
