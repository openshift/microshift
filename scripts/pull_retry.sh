#!/bin/bash
set -euo pipefail
set -x
#
# Use this script to avoid image pull getting stuck once MicroShift tries to
# do so in runtime.
#
PULL_CMD=${PULL_CMD:-"sudo crictl pull"}

function pull_image() {
    local -r image=$1
    local rc=0
    for _ in $(seq 3); do
        # shellcheck disable=SC2086
        timeout 5m ${PULL_CMD} "${image}" && return
        rc=$?
        sleep 1
    done
    exit ${rc}
}

if [ $# -lt 1 ] ; then
    echo "Usage: $(basename "$0") [images_to_download]..."
    exit 1
fi

for img in "$@"; do
    echo "Pulling ${img}"
    pull_image "${img}"
done
