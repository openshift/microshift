#!/bin/bash
#
# This script should be run on the hypervisor to set up an ephemeral
# Prometheus server for those tests that require metrics checking.

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

PROMETHEUS_DIR="${IMAGEDIR}/prometheus"

usage() {
    cat - <<EOF
${BASH_SOURCE[0]} (start|stop)

  -h           Show this help.

start: Start Prometheus.

stop: Stop Prometheus.

EOF
}

action_stop() {
    echo "Stopping Prometheus"
    podman stop prometheus > /dev/null || true
    podman rm --force prometheus > /dev/null || true
}

action_start() {
    mkdir -p "${PROMETHEUS_DIR}"

    PROM_CONFIG="${PROMETHEUS_DIR}/prometheus.yml"
    # Empty configuration file will take all defaults.
    # A config file is required to add remote-write enabling.
    touch "${PROM_CONFIG}"

    echo "Stopping previous instance of Prometheus"
    action_stop

    echo "Starting Prometheus"
    podman run -d --rm --name prometheus \
        -p 9091:9090 \
        -v "${PROMETHEUS_DIR}:/etc/prometheus" \
        docker.io/prom/prometheus \
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
