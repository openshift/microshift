#!/bin/bash
#
# This script should be run on the image building host to manage the
# number of workers for creating images in parallel.

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "${SCRIPTDIR}/common.sh"

if [ $# -eq 0 ]; then
    error "Usage: ${0} number-of-workers"
    exit 1
fi
num_workers="${1}"

# Loop from 2 because we should already have at least 1 worker.
for i in $(seq 2 "${num_workers}"); do
    sudo systemctl start "osbuild-worker@${i}.service"
done
