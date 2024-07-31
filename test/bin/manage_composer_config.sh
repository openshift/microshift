#!/bin/bash
#
# A script for managing osbuild-composer.

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

usage() {
    cat - <<EOF
${BASH_SOURCE[0]} (create [num_workers]|cleanup)

  -h           Show this help.

create: Set up system for building images, start webserver.
    [num_workers]: Number of workers to create for
                   building in parallel.

cleanup: Cancel any running builds, delete failed
         and completed builds, and remove package 
         sources other than the defaults.

EOF
}

action_create() {
    "${ROOTDIR}/scripts/image-builder/configure.sh"

    # Optionally create workers for building in parallel
    if [ $# -ne 0 ]; then
        create_workers "${1}"
    fi
    
    "${TEST_DIR}/bin/start_webserver.sh"
}

create_workers() {
    if [ $# -eq 0 ]; then
        usage
        exit 1
    fi
    workers="${1}"
    echo "Creating ${1} workers"

    # Loop from 2 because we should already have at least 1 worker.
    for i in $(seq 2 "${workers}"); do
        sudo systemctl start "osbuild-worker@${i}.service"
    done
}

action_cleanup() {
    # Clean up the composer cache
    "${ROOTDIR}/scripts/image-builder/cleanup.sh" -full

    "${TEST_DIR}/bin/start_webserver.sh stop"
}

if [ $# -eq 0 ]; then
    usage
    exit 1
fi
action="${1}"
shift

case "${action}" in
    create|cleanup)
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

