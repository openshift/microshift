#!/bin/bash
#
# This script should be run on the hypervisor to set up a caddy
# file-server for the images used by Vms running test scenarios.

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "${SCRIPTDIR}/common.sh"

echo "Starting web server in ${IMAGEDIR}"
cd "${IMAGEDIR}"

pkill caddy || true
nohup caddy file-server \
      --access-log \
      --browse \
      --listen "0.0.0.0:${WEB_SERVER_PORT}" \
      --root "${IMAGEDIR}" >caddy.log 2>&1 &
