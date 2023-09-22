#!/bin/bash
set -e -o pipefail

SCRIPT_NAME=$(basename "$0")

# Source the MicroShift health check functions library
# shellcheck source=packaging/greenboot/functions.sh
source /usr/share/microshift/functions/greenboot.sh

# Set the exit handler to log the exit status
trap 'script_exit' EXIT

# The script exit handler logging the FAILURE or FINISHED message depending
# on the exit status of the last command
#
# args: None
# return: None
function script_exit() {
    [ "$?" -ne 0 ] && echo "FAILURE" || echo "FINISHED"
}

#
# Main
#

# Exit if the current user is not 'root'
if [ "$(id -u)" -ne 0 ] ; then
    echo "The '${SCRIPT_NAME}' script must be run with the 'root' user privileges"
    exit 1
fi

echo "STARTED"

# Print the boot variable status
print_boot_status

# Exit normally if the boot counter is not set to zero
if ! grub2-editenv - list | grep -q ^boot_counter=0 ; then
    echo "The system is not scheduled to roll back on the next boot"
    exit 0
fi

echo "System rollback imminent"

echo "Instructing MicroShift to restore backup on rollback"
touch /var/lib/microshift-backups/restore

echo "Preparing MicroShift for a clean start"
microshift-cleanup-data --ovn
