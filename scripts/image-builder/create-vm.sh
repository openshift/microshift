#!/bin/bash
set -euo pipefail

if [ $# -ne 3 ] ; then
    echo "Usage: $(basename "$0") <vm_name> <network_name> <iso_path>"
    exit 1
fi

VMNAME=$1
NETNAME=$2
CDROM=$3

if [ ! -e "${CDROM}" ] ; then
    echo "The image ISO '${CDROM}' file does not exist."
    exit 1
fi

# Allow the VM creation to run in parallel using multiple instances of this script
# Note: If 'dnf' command is run in parallel, its database is corrupted
# === Start critical section ===
exec {LOCK_FD}<"$0"
flock --exclusive ${LOCK_FD}

sudo dnf install -y libvirt virt-manager virt-install virt-viewer libvirt-client qemu-kvm qemu-img sshpass
if [ "$(systemctl is-active libvirtd.socket)" != "active" ] ; then
    echo "Enabling libvirtd"
    sudo systemctl enable --now libvirtd
fi
# Necessary to allow remote connections in the virt-viewer application
sudo usermod -a -G libvirt "$(whoami)"

# === End critical section ===
flock --unlock ${LOCK_FD}

sudo bash -c " \
virt-install \
    --name ${VMNAME} \
    --vcpus 2 \
    --memory 3072 \
    --disk path=/var/lib/libvirt/images/${VMNAME}.qcow2,size=20 \
    --network network=${NETNAME},model=virtio \
    --events on_reboot=restart \
    --cdrom ${CDROM} \
    --noautoconsole \
    --wait \
"
