#!/bin/bash
#
# This script should be run on the hypervisor to set up libvirt and firewall.

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

usage() {
    cat - <<EOF
${BASH_SOURCE[0]} (create|cleanup)

  -h           Show this help.

create: Set up firewall, storage pool and network.
        Uses the VM_STORAGE_POOL, VM_DISK_DIR and
        VM_ISOLATED_NETWORK variables.

cleanup: Undo the settings made by 'create' command.

EOF
}

firewall_settings() {
    local -r action=$1

    for netname in default "${VM_ISOLATED_NETWORK}" ; do
        if ! sudo virsh net-info "${netname}" &>/dev/null ; then
            continue
        fi

        local vm_bridge
        local vm_bridge_cidr
        vm_bridge=$(sudo virsh net-info "${netname}" | grep '^Bridge:' | awk '{print $2}')
        vm_bridge_cidr=$(ip -f inet addr show "${vm_bridge}" | grep inet | awk '{print $2}')

        sudo firewall-cmd --permanent --zone=trusted "--${action}-source"="${vm_bridge_cidr}"
        sudo firewall-cmd --permanent --zone=public  "--${action}-port"="${WEB_SERVER_PORT}/tcp"
        sudo firewall-cmd --reload
    done
}

action_create() {
    # Storage pool
    if ! sudo virsh pool-info "${VM_STORAGE_POOL}" &>/dev/null; then
        sudo virsh pool-define-as "${VM_STORAGE_POOL}" dir --target "${VM_DISK_DIR}"
        sudo virsh pool-build "${VM_STORAGE_POOL}"
        sudo virsh pool-start "${VM_STORAGE_POOL}"
        sudo virsh pool-autostart "${VM_STORAGE_POOL}"
    fi

    # Isolated network
    if ! sudo sudo virsh net-info "${VM_ISOLATED_NETWORK}" &>/dev/null ; then
        local -r netconfig_tmpl="${SCRIPTDIR}/../assets/isolated-network.xml"
        local -r netconfig_file="${IMAGEDIR}/infra/isolated-network.xml"

        mkdir -p "$(dirname "${netconfig_file}")"
        envsubst <"${netconfig_tmpl}" >"${netconfig_file}"

        sudo virsh net-define    "${netconfig_file}"
        sudo virsh net-start     "${VM_ISOLATED_NETWORK}"
        sudo virsh net-autostart "${VM_ISOLATED_NETWORK}"
    fi

    # Firewall
    firewall_settings "add"
}

action_cleanup() {
    # Firewall part must run before the network configuration is undone
    firewall_settings "remove"

    # Isolated network
    if sudo virsh net-info "${VM_ISOLATED_NETWORK}" &>/dev/null ; then
        sudo virsh net-destroy "${VM_ISOLATED_NETWORK}"
        sudo virsh net-undefine "${VM_ISOLATED_NETWORK}"
    fi

    # Storage pool
    if sudo virsh pool-info "${VM_STORAGE_POOL}" &>/dev/null; then
        sudo virsh pool-destroy "${VM_STORAGE_POOL}"
        sudo virsh pool-undefine "${VM_STORAGE_POOL}"
    fi

}

if [ $# -eq 0 ]; then
    usage
    exit 1
fi
action="${1}"
shift

case "${action}" in
    create|cleanup)
        "action_${action}" "$@"
        ;;
    -h)
        usage
        exit 0
        ;;
    *)
        usage
        exit 1
        ;;
esac
