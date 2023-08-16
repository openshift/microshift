#!/bin/bash

# The location of the test directory, relative to the script.
TESTDIR="$(cd "${SCRIPTDIR}/.." && pwd)"

# The location of the root of the git repo, relative to the script.
ROOTDIR="$(cd "${TESTDIR}/.." && pwd)"

# The location of shared kickstart templates
# shellcheck disable=SC2034  # used elsewhere
KICKSTART_TEMPLATE_DIR="${TESTDIR}/kickstart-templates"

# The blueprint we should use for building an installer image.
# shellcheck disable=SC2034  # used elsewhere
INSTALLER_IMAGE_BLUEPRINT="rhel-9.2"

# The location for downloading all of the image-related output.
# The location the web server should serve.
export IMAGEDIR="${ROOTDIR}/_output/test-images"

# The storage pool name for VMs
VM_POOL_BASENAME="vm-storage"

# The location for storage for the VMs.
# shellcheck disable=SC2034  # used elsewhere
VM_DISK_BASEDIR="${IMAGEDIR}/${VM_POOL_BASENAME}"

# The isolated network name used by some VMs.
# shellcheck disable=SC2034  # used elsewhere
export VM_ISOLATED_NETWORK="isolated"

# Location of RPMs built from source
# shellcheck disable=SC2034  # used elsewhere
RPM_SOURCE="${ROOTDIR}/_output/rpmbuild"

# Location of RPMs built from source
# shellcheck disable=SC2034  # used elsewhere
NEXT_RPM_SOURCE="${ROOTDIR}/_output/rpmbuild-fake-next-minor"

# Location of RPMs built from source
# shellcheck disable=SC2034  # used elsewhere
YPLUS2_RPM_SOURCE="${ROOTDIR}/_output/rpmbuild-fake-yplus2-minor"

# Location of RPMs built from source
# shellcheck disable=SC2034  # used elsewhere
BASE_RPM_SOURCE="${ROOTDIR}/_output/rpmbuild-base"

# Location of local repository used by composer
# shellcheck disable=SC2034  # used elsewhere
LOCAL_REPO="${IMAGEDIR}/rpm-repos/microshift-local"

# Location of local repository used by composer
# shellcheck disable=SC2034  # used elsewhere
NEXT_REPO="${IMAGEDIR}/rpm-repos/microshift-fake-next-minor"

# Location of local repository used by composer
# shellcheck disable=SC2034  # used elsewhere
YPLUS2_REPO="${IMAGEDIR}/rpm-repos/microshift-fake-yplus2-minor"

# Location of local repository used by composer
# shellcheck disable=SC2034  # used elsewhere
BASE_REPO="${IMAGEDIR}/rpm-repos/microshift-base"

# Location of data files created by the tools for managing scenarios
# as they are run.
#
# The CI system will override this, but we need a default for local
# use. Use the image directoy, since that is already served by a web
# server.
#
# shellcheck disable=SC2034  # used elsewhere
SCENARIO_INFO_DIR="${SCENARIO_INFO_DIR:-${IMAGEDIR}/scenario-info}"

# Directory to crawl for scenarios when creating/running in batch mode.
#
# The CI system will override this depending on the job its running.
SCENARIO_SOURCES="${SCENARIO_SOURCES:-${TESTDIR}/scenarios}"

# The location of the robot framework virtualenv.
# The CI system will override this.
# shellcheck disable=SC2034  # used elsewhere
RF_VENV=${RF_VENV:-${ROOTDIR}/_output/robotenv}

# Which port the web server should run on.
WEB_SERVER_PORT=${WEB_SERVER_PORT:-8080}

title() {
    # Only use color when reporting to a terminal
    if [ -t 1 ]; then
        echo -e "\E[34m\n# $1\E[00m"
    else
        echo "$1"
    fi
}

error() {
    local message="$*"
    echo "ERROR: ${message} [$(caller)]" 1>&2
}

get_vm_bridge_interface() {
    local -r netdevice=${1:-default}

    # Return an empty string if virsh is not installed
    if ! which virsh &>/dev/null ; then
        echo ""
        return
    fi
    # Return an empty string if the network does not exist
    if ! sudo virsh net-info "${netdevice}" &>/dev/null ; then
        echo ""
        return
    fi

    # $ sudo virsh net-info default
    # Name:           default
    # UUID:           eaac9592-2324-4ae6-b2ec-a5ae94272456
    # Active:         yes
    # Persistent:     yes
    # Autostart:      yes
    # Bridge:         virbr0
    sudo virsh net-info "${netdevice}" | grep '^Bridge:' | awk '{print $2}'
}

get_vm_bridge_ip() {
    local -r netdevice=${1:-default}
    local -r bridge="$(get_vm_bridge_interface "${netdevice}")"

    # When get_vm_bridge_interface is run on the CI cluster there is
    # no bridge, and possibly no virsh command. Do not fail, but
    # return an empty IP.
    if [ -z "${bridge}" ]; then
        echo ""
        return
    fi

    # $ ip -f inet addr show virbr0
    # 10: virbr0: <NO-CARRIER,BROADCAST,MULTICAST,UP> mtu 1500 qdisc noqueue state DOWN group default qlen 1000
    #     inet 192.168.122.1/24 brd 192.168.122.255 scope global virbr0
    #        valid_lft forever preferred_lft forever
    ip -f inet addr show "${bridge}" | grep inet | awk '{print $2}' | cut -d/ -f1
}

# The IP address of the current host on the bridge used for the
# default network for libvirt VMs.
# shellcheck disable=SC2034  # used elsewhere
VM_BRIDGE_IP="$(get_vm_bridge_ip "default")"
