#!/bin/bash
set -e -o pipefail

SCRIPT_NAME=$(basename $0)

# Set the exit handler to log the exit status
trap 'script_exit' EXIT

# The script exit handler logging the FAILURE or FINISHED message depending
# on the exit status of the last command
#
# args: None
# return: None
function script_exit() {
    [ "$?" -ne 0 ] && status=FAILURE || status=FINISHED
    echo $status
}

#
# Main
#

# Exit if the current user is not 'root'
if [ $(id -u) -ne 0 ] ; then
    echo "The '${SCRIPT_NAME}' script must be run with the 'root' user privileges"
    exit 1
fi

echo "STARTED"

# Dump GRUB boot variables and ostree status affecting the script behavior.
# This information is important for troubleshooting cleanup issues.
grub_vars=$(grub2-editenv - list | grep ^boot_ || true)
[ -z "${grub_vars}" ] && grub_vars=None

echo -e "GRUB boot variables:\n${grub_vars}"
echo -e "The ostree status:\n$(ostree admin status || true)"

# Exit normally if the boot counter is not set to zero
if ! grub2-editenv - list | grep -q ^boot_counter=0 ; then
    echo "The system is not scheduled to roll back on the next boot"
    exit 0
fi

echo "System rollback imminent - preparing MicroShift for a clean start"
microshift-cleanup-data --ovn
