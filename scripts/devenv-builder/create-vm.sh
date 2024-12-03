#!/bin/bash
set -euo pipefail

ROOTDIR=$(git rev-parse --show-toplevel)/scripts/devenv-builder

function usage() {
    echo "Usage: $(basename "$0") [<VMNAME> <VMDISKDIR> <ISOFILE> <NCPUS> <RAMSIZE> <DISKSIZE> <SWAPSIZE> <DATAVOLSIZE>]"
    echo "INFO: Specify 0 swap size to disable swap partition"
    echo "INFO: Positional arguments also can be specified using environment variables"
    echo "INFO: All sizes in GB"
    [ -n "$1" ] && echo -e "\nERROR: $1"
    exit 1
}

VMNAME=${1:-${VMNAME}}
VMDISKDIR=${2:-${VMDISKDIR}}
ISOFILE=${3:-${ISOFILE}}
NCPUS=${4:-${NCPUS}}
RAMSIZE=${5:-${RAMSIZE}}
DISKSIZE=${6:-${DISKSIZE}}
SWAPSIZE=${7:-${SWAPSIZE}}
DATAVOLSIZE=${8:-${DATAVOLSIZE}}
[ -z "${VMNAME}" ]      && usage "Invalid VM name: '${VMNAME}'"
[ ! -e "${VMDISKDIR}" ] && usage "VM disk directory '${VMDISKDIR}' is not accessible"
[ ! -e "${ISOFILE}" ]   && usage "Installation ISO file '${ISOFILE}' is not accessible"

[[ ! "${NCPUS}" =~ ^[0-9]+$ ]]       || [[ "${NCPUS}" -le 0 ]]       && usage "Invalid number of CPUs: '${NCPUS}'"
[[ ! "${RAMSIZE}" =~ ^[0-9]+$ ]]     || [[ "${RAMSIZE}" -le 0 ]]     && usage "Invalid RAM size: '${RAMSIZE}'"
[[ ! "${DISKSIZE}" =~ ^[0-9]+$ ]]    || [[ "${DISKSIZE}" -le 0 ]]    && usage "Invalid disk size: '${DISKSIZE}'"
[[ ! "${SWAPSIZE}" =~ ^[0-9]+$ ]]    || [[ "${SWAPSIZE}" -lt 0 ]]    && usage "Invalid swap size: '${SWAPSIZE}'"
[[ ! "${DATAVOLSIZE}" =~ ^[0-9]+$ ]] || [[ "${DATAVOLSIZE}" -le 0 ]] && usage "Invalid data volume size: '${DATAVOLSIZE}'"

# RAM size is expected in MB
RAMSIZE=$(( RAMSIZE * 1024 ))
# Calculate system root partition size (1GB is allocated to the boot partition)
SYSROOTSIZE=$(( DISKSIZE - 1 - SWAPSIZE - DATAVOLSIZE ))
# System root size is expected in MB
SYSROOTSIZE=$(( SYSROOTSIZE * 1024 ))
# Swap size is expected in MB
SWAPSIZE=$(( SWAPSIZE * 1024 ))
# Network name
NETWORK=${NETWORK:-default}
# Pool name - determin the pool for vm disk the volume
MICROSHIFT_VOL_POOL="${MICROSHIFT_VOL_POOL:-default}"


KICKSTART_FILE=$(mktemp "/tmp/kickstart-${VMNAME}-XXXXX.ks")
cat < "${ROOTDIR}/config/kickstart.ks.template" | \
    sed "s;REPLACE_HOST_NAME;${VMNAME};" | \
    sed "s;REPLACE_SWAP_SIZE;${SWAPSIZE};" | \
    sed "s;REPLACE_LVM_SYSROOT_SIZE;${SYSROOTSIZE};" > "${KICKSTART_FILE}"
# Disable swap if its size is 0
if [ "${SWAPSIZE}" -eq 0 ] ; then
    sed -i "s;^part swap;#part swap;" "${KICKSTART_FILE}"
fi

sudo bash -c " \
cd ${VMDISKDIR} && \
virt-install \
    --name ${VMNAME} \
    --vcpus ${NCPUS} \
    --memory ${RAMSIZE} \
    --disk pool=${MICROSHIFT_VOL_POOL},path=./${VMNAME}.qcow2,size=${DISKSIZE} \
    --network network=${NETWORK},model=virtio \
    --events on_reboot=restart \
    --location ${ISOFILE} \
    --initrd-inject=${KICKSTART_FILE} \
    --extra-args \"inst.ks=file:/$(basename "${KICKSTART_FILE}")\" \
    --graphics vnc,listen=0.0.0.0 \
    --noautoconsole \
    --wait \
"
