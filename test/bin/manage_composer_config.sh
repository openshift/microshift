#!/bin/bash
#
# A script for managing osbuild-composer.

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

usage() {
    cat - <<EOF
${BASH_SOURCE[0]} (workers <number>|cleanup)

  -h           Show this help.

workers: Create multiple workers for building images
         in parallel.
    <number>: Number of workers to create.

cleanup: Cancel any running builds, delete failed
         and completed builds, and remove package 
         sources other than the defaults.

EOF
}

action_workers() {
    if [ $# -eq 0 ]; then
        usage
        exit 1
    fi
    workers="${1}"

    # Loop from 2 because we should already have at least 1 worker.
    for i in $(seq 2 "${num_workers}"); do
        sudo systemctl start "osbuild-worker@${i}.service"
    done
}

action_cleanup() {
    # Clean up the composer cache
    "${ROOTDIR}/scripts/image-builder/cleanup.sh" -full
}

if [ $# -eq 0 ]; then
    usage
    exit 1
fi
action="${1}"
shift

case "${action}" in
    workers|cleanup)
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

