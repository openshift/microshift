#!/bin/bash
set -e

ROOTDIR=$(git rev-parse --show-toplevel)
ISODIR=${ROOTDIR}/_output/image-builder

if [ $# -ne 2 ] ; then
    echo "Usage: $(basename "$0") <vm_name> <network_name>"
    exit 1
fi

VMNAME=$1
NETNAME=$2
CDROM=$(ls -1 "${ISODIR}/microshift-installer-*.$(uname -i).iso" 2>/dev/null || true)

if [ ! -e "${CDROM}" ] ; then
    echo "The image ISO '${CDROM}' file does not exist. Run 'make iso' to create it"
    exit 1
fi

sudo dnf install -y libvirt virt-manager virt-install virt-viewer libvirt-client qemu-kvm qemu-img sshpass
if [ "$(systemctl is-active libvirtd.socket)" != "active" ] ; then
    echo "Restart your host to initialize the virtualization environment"
    exit 1
fi
# Necessary to allow remote connections in the virt-viewer application
sudo usermod -a -G libvirt "$(whoami)"

sudo -b bash -c " \
virt-install \
    --name ${VMNAME} \
    --vcpus 2 \
    --memory 3072 \
    --disk path=/var/lib/libvirt/images/${VMNAME}.qcow2,size=20 \
    --network network=${NETNAME},model=virtio \
    --os-type generic \
    --events on_reboot=restart \
    --cdrom ${CDROM} \
    --noautoconsole \
    --wait \
"
