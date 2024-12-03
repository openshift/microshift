#!/usr/bin/bash

set -eo pipefail

# https://github.com/openshift/microshift/blob/main/docs/devenv_setup.md
# https://github.com/openshift/microshift/blob/main/docs/devenv_setup_auto.md

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPOROOT="$( cd "${SCRIPTDIR}/../.." && pwd )"

# Establish defaults
MICROSHIFT_VMDISKDIR="${MICROSHIFT_VMDISKDIR:-/var/lib/libvirt/images}"
MICROSHIFT_SSH_KEY_FILE="${MICROSHIFT_SSH_KEY_FILE:-${HOME}/.ssh/id_rsa.pub}"
MICROSHIFT_RHEL_VERSION="${MICROSHIFT_RHEL_VERSION:-9}"
MICROSHIFT_VOL_POOL="${MICROSHIFT_VOL_POOL:-default}"

# We need to pass these parameters to ssh when trying to use sshpass
# to ensure that we do not get authentication errors because too many
# ssh keys are tried before defaulting to password authentication.
SSH_PASSWORD_OPTS="-o PubkeyAuthentication=no -o PreferredAuthentications=password -o StrictHostKeyChecking=no"

# Show the IP address of the VM
function get_ip {
    sudo virsh domifaddr "$1" \
        | grep ipv \
        | awk '{print $4}' \
        | cut -f1 -d/
}

# Use the RHEL version and other settings to build a unique VM name
# that will be discoverable in the portal in case the subscription
# needs to be managed by hand.
function get_vm_name {
    local rhel_version="$1"

    echo "microshift-${USER}-$(hostname -s)-${rhel_version}"
}

# Use the RHEL version to get the name of the ISO file to use for
# installing the VM.
function get_base_isofile {
    local rhel_version="$1"

    case ${rhel_version} in
        8)
            echo "rhel-8.10-$(uname -m)-dvd.iso"
            ;;
        8.*)
            echo "rhel-${rhel_version}-$(uname -m)-dvd.iso"
            ;;
        9)
            echo "rhel-9.4-$(uname -m)-dvd.iso"
            ;;
        9.*)
            echo "rhel-${rhel_version}-$(uname -m)-dvd.iso"
            ;;
        *)
            echo "Unknown RHEL version ${rhel_version}" 1>&2
            exit 1
    esac
}

function action_config() {
    local -r deps="libvirt virt-manager virt-install virt-viewer libvirt-client qemu-kvm qemu-img sshpass wget"
    
    "${SCRIPTDIR}/../dnf_retry.sh" "install" "${deps}"

    if [ "$(systemctl is-active libvirtd.socket)" != "active" ] ; then
        echo "Enabling libvirtd"
        sudo systemctl enable --now libvirtd
    fi
    # Necessary to allow remote connections in the virt-viewer application
    sudo usermod -a -G libvirt "$(whoami)"

    binary="yq_linux_amd64"
    if [ "$(arch)" == "aarch64" ]; then
        binary="yq_linux_arm64"
    fi

    sudo wget https://github.com/mikefarah/yq/releases/latest/download/"${binary}" -O /usr/bin/yq && sudo chmod +x /usr/bin/yq
}

# Create the VM, if it does not exist
function action_create {
    # Set the variables needed by create-vm.sh instead of passing them as
    # command line arguments.
    export VMDISKDIR="${MICROSHIFT_VMDISKDIR}"
    export NCPUS="${NCPUS:-4}"
    export RAMSIZE="${RAMSIZE:-8}"
    export DISKSIZE="${DISKSIZE:-100}"
    export SWAPSIZE="${SWAPSIZE:-8}"
    export DATAVOLSIZE="${DATAVOLSIZE:-2}"
    export MICROSHIFT_VOL_POOL="${MICROSHIFT_VOL_POOL}"
    if [ -z "${ISOFILE}" ]; then
        ISOFILE="${VMDISKDIR}/$(get_base_isofile "${MICROSHIFT_RHEL_VERSION}")"
    fi
    export ISOFILE

    # fail if cd fails
    cd "${REPOROOT}" || (echo "could not cd into repo root(${REPOROOT})" 1>&2 && exit 1)

    if [ ! -d "${VMDISKDIR}" ]; then
        echo "Creating ${VMDISKDIR} ..."
        mkdir -p "${VMDISKDIR}"
    fi

    if ! sudo virsh desc "${VMNAME}" >/dev/null 2>&1; then
        echo "Creating VM ${VMNAME} from ${ISOFILE} ..."
        if ! ./scripts/devenv-builder/create-vm.sh; then
            echo "exiting failure" 1>&2
            exit 1
        fi
    else
        echo "VM ${VMNAME} exists"
    fi

    # Wait for an IP to be assigned
    ip=$(get_ip "${VMNAME}")
    while [ -z "${ip}" ]; do
        echo "Waiting for VM ${VMNAME} to have an IP"
        sleep 30
        ip=$(get_ip "${VMNAME}")
    done

    # Remove any old keys for that IP
    echo "Removing old ssh keys for ${ip} ..."
    ssh-keygen -R "${ip}"

    # Wait for sshd to be available
    # shellcheck disable=SC2086
    while ! sshpass -p microshift ssh ${SSH_PASSWORD_OPTS} "microshift@${ip}" true; do
        echo "Waiting to be able to login to microshift@${ip}..."
        sleep 30
    done

    # Copy ssh key into the host for passwordless access
    echo "Copying ssh key ${MICROSHIFT_SSH_KEY_FILE} ..."
    # shellcheck disable=SC2086
    sshpass -p microshift \
            ssh-copy-id \
            -f \
            -i "${MICROSHIFT_SSH_KEY_FILE}" \
            ${SSH_PASSWORD_OPTS} \
            "microshift@${ip}"

    echo "VM online"

    # Initialize RH subscription
    echo "Checking subscription-manager..."
    ssh -t "microshift@${ip}" "if ! sudo subscription-manager status; then sudo subscription-manager register --auto-attach; fi"

    echo "VM online at ${ip}"
}

function action_ssh {
    ip=$(get_ip "${VMNAME}")

    if [ -z "${ip}" ]; then
        echo "${VMNAME} has no IP, try the 'console' command to this script"
        return 1
    fi
    ssh "microshift@${ip}"
}

function action_console {
    echo
    echo "After connected, press RETURN to generate a login prompt and Control-] to disconnect"
    echo
    sudo virsh console "${VMNAME}"
}

function action_delete {
    ip=$(get_ip "${VMNAME}")

    if [ -n "${ip}" ]; then
        ssh "microshift@${ip}" "sudo subscription-manager unregister" || true
    fi

    sudo virsh destroy "${VMNAME}"
    sudo virsh undefine --nvram "${VMNAME}"

    # FIXME: The volume pool here may not be standard. How do we
    # figure out what it should be?
    sudo virsh vol-delete "${VMNAME}.qcow2" "${MICROSHIFT_VOL_POOL}"
}

function action_rm {
    action_delete
}

function action_ip {
    if ! sudo virsh desc "${VMNAME}" >/dev/null 2>&1; then
        echo "VM ${VMNAME} does not exist" 1>&2
        exit 1
    fi

    get_ip "${VMNAME}"
}

function action_help {
    usage
}

function usage {
    local -r script_name=$(basename "$0")
    cat - <<EOF
${script_name} (config|create|ip|ssh|console|delete|rm|help) [options]

Commands:

  config    -- install and start libvirt
  create    -- create a new VM
  ip        -- show the IP of the VM
  ssh       -- ssh into the VM
  console   -- connect to the serial console of the VM
  delete|rm -- delete the VM
  help      -- show this help

Options:

  -d MICROSHIFT_VMDISKDIR  Specify the location the VM
                           disk(s) should be created
                           (Default: ${MICROSHIFT_VMDISKDIR})

  -i ISOFILE  Specify the location of the RHEL
              ISO file instead of relying on computing
              the name from the RHEL version and VMDISKDIR.
              (Default: ${MICROSHIFT_VMDISKDIR}/$(get_base_isofile "${MICROSHIFT_RHEL_VERSION}"))

  -k MICROSHIFT_SSH_KEY_FILE  Specify the location of the
                              SSH public key to install into
                              the image.
                              (Default: ${MICROSHIFT_SSH_KEY_FILE})

  -n VMNAME  Specify the name for the new VM
             (Default: $(get_vm_name "${MICROSHIFT_RHEL_VERSION}"))

  -v MICROSHIFT_RHEL_VERSION  Specify the RHEL version to default to.
                              Can be a major version ("9") or a
                              major and minor version ("8.7").
                              (Default: ${MICROSHIFT_RHEL_VERSION})

Variables:

  MICROSHIFT_VMDISKDIR -- Default location for VM disk files.

  MICROSHIFT_VOL_POOL -- Default libvirt volume pool associated with
  VMDISKDIR.

  MICROSHIFT_SSH_KEY_FILE -- Default ssh public key file to install.

  MICROSHIFT_RHEL_VERSION -- Default RHEL version to use.

EOF
}

# The first argument should always be what action to take.
case "$1" in
    config)
        action_config
        exit 0
        ;;
    create|ip|ssh|console|delete|rm|help)
        action="$1"
        shift
        ;;
    *)
        usage
        exit 1
        ;;
esac

# The other arguments are all the same, regardless of the action.
while getopts "d:i:k:n:v:" opt; do
    case ${opt} in
        d)
            MICROSHIFT_VMDISKDIR="${OPTARG}"
            ;;
        i)
            ISOFILE="${OPTARG}"
            ;;
        k)
            MICROSHIFT_SSH_KEY_FILE="${OPTARG}"
            ;;
        n)
            VMNAME="${OPTARG}"
            ;;
        v)
            MICROSHIFT_RHEL_VERSION="${OPTARG}"
            ;;
        *)
            usage
            exit 1
            ;;
    esac
done
shift $((OPTIND-1))

# We always need to know the VMNAME, so compute it one time.
if [ -z "${VMNAME}" ]; then
    VMNAME=$(get_vm_name "${MICROSHIFT_RHEL_VERSION}")
fi
export VMNAME

"action_${action}"
