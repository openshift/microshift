#!/bin/bash
#
# This script should be run on the hypervisor to set up an ephemeral
# Loki server for those tests that require metrics checking.

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

usage() {
    cat - <<EOF
${BASH_SOURCE[0]} (start|stop)

  -h           Show this help.

start: Start Loki.

stop: Stop Loki.

EOF
}

action_stop() {
    echo "Stopping Loki"
    podman stop loki > /dev/null || true
    podman rm --force loki > /dev/null || true
}

action_start() {
    echo "Stopping previous instance of Loki"
    action_stop

    echo "Starting Loki"
    podman run -d --rm --name loki \
        -p 3100:3100 \
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
