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

# The location for storage for the VMs.
# shellcheck disable=SC2034  # used elsewhere
VM_DISK_DIR="${IMAGEDIR}/vm-storage"

# Location of RPMs built from source
# shellcheck disable=SC2034  # used elsewhere
RPM_SOURCE="${ROOTDIR}/_output/rpmbuild"

# Location of local repository used by composer
# shellcheck disable=SC2034  # used elsewhere
LOCAL_REPO="${IMAGEDIR}/microshift-local"

# Location of data files created by the tools for managing scenarios
# as they are run.
#
# The CI system will override this, but we need a default for local
# use. Use the image directoy, since that is already served by a web
# server.
#
# shellcheck disable=SC2034  # used elsewhere
SCENARIO_INFO_DIR="${SCENARIO_INFO_DIR:-${IMAGEDIR}/scenario-info}"

# The location of the robot framework virtualenv.
# The CI system will override this.
# shellcheck disable=SC2034  # used elsewhere
RF_VENV=${RF_VENV:-${ROOTDIR}/_output/robotenv}

# Which port the web server should run on.
WEB_SERVER_PORT=${WEB_SERVER_PORT:-8080}

error() {
    local message="$*"
    echo "ERROR: ${message} [$(caller)]" 1>&2
}

get_vm_bridge_interface() {
    # $ sudo virsh net-info default
    # Name:           default
    # UUID:           eaac9592-2324-4ae6-b2ec-a5ae94272456
    # Active:         yes
    # Persistent:     yes
    # Autostart:      yes
    # Bridge:         virbr0

    sudo virsh net-info default | grep '^Bridge:' | awk '{print $2}'
}

get_vm_bridge_ip() {
    local bridge

    # When get_vm_bridge_interface is run on the CI cluster there is
    # no bridge, and possibly no virsh command. Do not fail, but
    # return an empty IP.
    bridge="$(get_vm_bridge_interface || true)"

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
VM_BRIDGE_IP="$(get_vm_bridge_ip)"
