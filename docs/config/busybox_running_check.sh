#!/bin/bash
set -e

SCRIPT_NAME=$(basename $0)

# Source the MicroShift health check functions library
source /usr/share/microshift/functions/greenboot.sh

# Exit if the current user is not 'root'
if [ $(id -u) -ne 0 ] ; then
    echo "The '${SCRIPT_NAME}' script must be run with the 'root' user privileges"
    exit 1
fi

echo "STARTED"

# Set the wait timeout for the current check based on the boot counter
WAIT_TIMEOUT_SECS=$(get_wait_timeout)

/usr/bin/microshift healthcheck -v=2 --timeout="${WAIT_TIMEOUT_SECS}s" --namespace busybox --deployments busybox-deployment
