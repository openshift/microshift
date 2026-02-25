#!/usr/bin/env bash

set -xeuo pipefail

# boot_success=0 is set when deployment is staged or when grub boots the system.
# boot_success=1 is set when greenboot succeeds after deploying new image.

# When the deployment testing is done, boot_counter should be removed.
if grub2-editenv - list | grep -q ^boot_counter=; then
    echo "Greenboot didn't decide the system is healthy after staging new deployment."
    echo "Quitting to not interfere with the process"
    exit 0
fi

echo "System is unhealthy and greenboot's 'deployment testing' procedure is not active - running auto-recovery for MicroShift"

echo "Making sure MicroShift is stopped and doesn't restart preventing from restoring the backup."
systemctl stop microshift

microshift restore --auto-recovery /var/lib/microshift-auto-recovery
systemctl reboot
