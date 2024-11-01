#!/bin/bash
#
# A script for managing osbuild-composer.

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

usage() {
    cat - <<EOF
${BASH_SOURCE[0]} (create|cleanup|create-workers [num_workers])

  -h           Show this help.

create: Set up system for building images, start webserver.

cleanup: Cancel any running builds, delete failed
         and completed builds, and remove package 
         sources other than the defaults.

create-workers: Create multiple osbuild workers for 
                building in parallel.
    [num_workers]: Number of workers. If unspecified,
                   The optimal number of workers will
                   be determined based on the number
                   of CPU cores.

EOF
}

action_create() {
    if [ $# -ne 0 ]; then
        usage
        exit 1
    fi
    
    "${ROOTDIR}/scripts/devenv-builder/configure-composer.sh"
    
    "${TESTDIR}/bin/manage_webserver.sh" "start"
}

action_create-workers() {
    local workers
    # If no number is given, determine the optimal number of workers
    if [ $# -eq 0 ]; then
        local -r cpu_cores="$(grep -c ^processor /proc/cpuinfo)"
        local -r max_workers=$(find "${ROOTDIR}/test/image-blueprints" -name \*.toml | wc -l)
        local -r cur_workers="$( [ "${cpu_cores}" -lt  $(( max_workers * 2 )) ] && echo $(( cpu_cores / 2 )) || echo "${max_workers}" )"

        workers="${cur_workers}"
    # If specified explicitly, create the given number of workers
    else
        workers="${1}"
    fi

    echo "Creating ${workers} workers"

    # Loop from 2 because we should already have at least 1 worker.
    for i in $(seq 2 "${workers}"); do
        sudo systemctl start "osbuild-worker@${i}.service"
    done
}

action_cleanup() {
    # Clean up the composer cache
    "${ROOTDIR}/scripts/devenv-builder/cleanup-composer.sh" -full

    "${TESTDIR}/bin/manage_webserver.sh" "stop"
}

if [ $# -eq 0 ]; then
    usage
    exit 1
fi
action="${1}"
shift

case "${action}" in
    create|cleanup|create-workers)
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

