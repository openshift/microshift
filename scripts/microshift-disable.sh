#!/bin/bash

set -e -o pipefail

SCRIPT_NAME=$(basename $0)

# Exit if the current user is not 'root'
if [ $(id -u) -ne 0 ] ; then
    echo "The '${SCRIPT_NAME}' script must be run with the 'root' user privileges"
    exit 1
fi

echo "Stopping MicroShift..."
systemctl stop microshift

echo "Stopping pods managed by kubelet..."
systemctl stop kubepods.slice

echo "Removing OVS configuration..."
/usr/bin/configure-ovs.sh OpenShiftSDN
