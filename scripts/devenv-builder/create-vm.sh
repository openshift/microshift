#!/bin/bash
#
# This script automates the VM creation steps described in the "MicroShift Development Environment on RHEL 8" document.
# See https://github.com/ggiguash/microshift/blob/main/docs/devenv_rhel8.md#creating-vm
#
set -eo pipefail
ROOTDIR=$(git rev-parse --show-toplevel)/scripts/devenv-builder

function usage() {
    echo "Usage: $(basename $0) <vm_name> <vm_disk_dir> <rhel_dvd_iso_file> <ncpus> <memory_in_GB> <disk_in_GB> <data_vol_size_in_GB>" 
    [ ! -z "$1" ] && echo -e "\nERROR: $1"
    exit 1
}

if [ $# -ne 7 ] ; then
    usage "Invalid number of arguments"
fi

VMNAME=$1
VMDISKDIR=$2
RHELISO=$3
NCPUS=$4
RAMSIZE=$5
DISKSIZE=$6
DATAVOLSIZE=$7
[ -z "${VMNAME}" ]                   && usage "Invalid VM name: '${VMNAME}'"
[ ! -e "${VMDISKDIR}" ]              && usage "VM disk directory '${VMDISKDIR}' is not accessible"
[ ! -e "${RHELISO}" ]                && usage "RHEL ISO file '${RHELISO}' is not accessible"
[[ ! "${NCPUS}" =~ ^[0-9]+$ ]]       && usage "Invalid number of CPUs: '${NCPUS}'"
[[ ! "${RAMSIZE}" =~ ^[0-9]+$ ]]     && usage "Invalid RAM size: '${RAMSIZE}'"
[[ ! "${DISKSIZE}" =~ ^[0-9]+$ ]]    && usage "Invalid disk size: '${DISKSIZE}'"
[[ ! "${DATAVOLSIZE}" =~ ^[0-9]+$ ]] && usage "Invalid data volume size: '${DATAVOLSIZE}'"

# RAM size is expected in MB
RAMSIZE=$(( ${RAMSIZE} * 1024 ))
# Calculate system root partition size (1GB is allocated to the boot partition)
SYSROOTSIZE=$(( ${DISKSIZE} - 1 - ${DATAVOLSIZE} ))
# System root size is expected in MB
SYSROOTSIZE=$(( ${SYSROOTSIZE} * 1024 ))

KICKSTART_FILE=/tmp/devenv-kickstart.ks
cat ${ROOTDIR}/config/kickstart.ks.template | \
    sed "s;REPLACE_HOST_NAME;${VMNAME};" | \
    sed "s;REPLACE_LVM_SYSROOT_SIZE;${SYSROOTSIZE};" > ${KICKSTART_FILE}

sudo dnf install -y libvirt virt-manager virt-viewer libvirt-client qemu-kvm qemu-img
sudo -b bash -c " \
cd ${VMDISKDIR} && \
virt-install \
    --name ${VMNAME} \
    --vcpus ${NCPUS} \
    --memory ${RAMSIZE} \
    --disk path=./${VMNAME}.qcow2,size=${DISKSIZE} \
    --network network=default,model=virtio \
    --events on_reboot=restart \
    --location ${RHELISO} \
    --initrd-inject=${KICKSTART_FILE} \
    --extra-args \"inst.ks=file:/$(basename ${KICKSTART_FILE})\" \
"
