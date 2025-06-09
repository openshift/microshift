#!/bin/bash
#
# This script should be run on the hypervisor to set up an ephemeral
# Prometheus server for those tests that require metrics checking.

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

PROMETHEUS_DIR="${IMAGEDIR}/prometheus"
DEFAULT_HOST_PORT="9092"

usage() {
    cat - <<EOF
${BASH_SOURCE[0]} (start|stop) [port]

  -h           Show this help.

start [port]: Start Prometheus.
             Uses port ${DEFAULT_HOST_PORT} on the host by default.
             The container name will be prometheus-<host_port>.

stop [port]: Stop Prometheus.
            Uses port ${DEFAULT_HOST_PORT} by default to identify the container.
            The container name is assumed to be prometheus-<host_port>.

EOF
}

action_stop() {
    local host_port="${1:-${DEFAULT_HOST_PORT}}"
    local container_name="prometheus-${host_port}"

    echo "Stopping Prometheus container ${container_name}"
    podman stop "${container_name}" > /dev/null || true
    podman rm --force "${container_name}" > /dev/null || true
}

action_start() {
    local host_port="${1:-${DEFAULT_HOST_PORT}}"
    local container_name="prometheus-${host_port}"

    mkdir -p "${PROMETHEUS_DIR}"
    PROM_CONFIG="${PROMETHEUS_DIR}/prometheus.yml"
    # Empty configuration file will take all defaults.
    # A config file is required to add remote-write enabling.
    # This file may be shared across any number of prometheus instances.
    touch "${PROM_CONFIG}"

    echo "Stopping previous instance of Prometheus container ${container_name} (if any)"
    action_stop "${host_port}"

    echo "Starting Prometheus container ${container_name} on host port ${host_port}"
    podman run -d --rm --name "${container_name}" \
        -p "${host_port}:9090" \
        -v "${PROMETHEUS_DIR}:/etc/prometheus:Z" \
        quay.io/prometheus/prometheus \
        --config.file=/etc/prometheus/prometheus.yml \
        --web.enable-remote-write-receiver > /dev/null
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
