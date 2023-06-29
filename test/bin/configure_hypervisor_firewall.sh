#!/bin/bash
#
# This script should be run on the hypervisor to configure the
# firewall for incoming connections to the VM network.

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "${SCRIPTDIR}/common.sh"

VM_BRIDGE=$(sudo virsh net-info default | grep '^Bridge:' | awk '{print $2}')
VM_BRIDGE_CIDR=$(ip -f inet addr show "${VM_BRIDGE}" | grep inet | awk '{print $2}')
sudo firewall-cmd --permanent --zone=trusted --add-source="${VM_BRIDGE_CIDR}"
sudo firewall-cmd --permanent --zone=public --add-port="${WEB_SERVER_PORT}/tcp"
sudo firewall-cmd --reload
