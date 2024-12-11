#!/bin/bash
#
# MicroShift OLM-specific functionality used in Greenboot health check procedures.
#
# If 'microshift-olm' RPM is installed, health check needs to include resources
# from the 'openshift-operator-lifecycle-manager' namespace.
#
set -eu -o pipefail

SCRIPT_NAME=$(basename "$0")

# Source the MicroShift health check functions library
# shellcheck source=packaging/greenboot/functions.sh
source /usr/share/microshift/functions/greenboot.sh

# Exit if the current user is not 'root'
if [ "$(id -u)" -ne 0 ] ; then
    echo "The '${SCRIPT_NAME}' script must be run with the 'root' user privileges"
    exit 1
fi

echo "STARTED"

# Print the boot variable status
print_boot_status

# Set the wait timeout for the current check based on the boot counter
WAIT_TIMEOUT_SECS=$(get_wait_timeout)

/usr/bin/microshift healthcheck -v=2 --timeout="${WAIT_TIMEOUT_SECS}s" --namespaces openshift-operator-lifecycle-manager
