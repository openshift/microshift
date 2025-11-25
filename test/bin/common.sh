#!/bin/bash

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    echo "This script must be sourced, not executed."
    exit 1
fi

UNAME_M=$(uname -m)
export UNAME_M

# The location of the test directory, relative to the script.
TESTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# The location of the root of the git repo, relative to the script.
ROOTDIR="$(cd "${TESTDIR}/.." && pwd)"

# Most output should be written under this directory
OUTPUTDIR="${ROOTDIR}/_output"

# The location of shared kickstart templates
# shellcheck disable=SC2034  # used elsewhere
KICKSTART_TEMPLATE_DIR="${TESTDIR}/kickstart-templates"

# The location for downloading all of the image-related output.
# The location the web server should serve.
export IMAGEDIR="${OUTPUTDIR}/test-images"

# Ginkgo test binary path
export GINKGO_TEST_BINARY="${OUTPUTDIR}/bin/extended-platform-tests"
export HANDLERESULT_SCRIPT="${OUTPUTDIR}/bin/handleresult.py"

# The storage pool base name for VMs.
# The actual pool names will be '${VM_POOL_BASENAME}-${SCENARIO}'.
VM_POOL_BASENAME="vm-storage"

# The location for storage for the VMs.
export VM_DISK_BASEDIR="${IMAGEDIR}/${VM_POOL_BASENAME}"

# The isolated network name used by some VMs.
export VM_ISOLATED_NETWORK="isolated"

# Libvirt network for Multus tests
export VM_MULTUS_NETWORK="multus"

# Libvirt network for IPv6 tests
export VM_IPV6_NETWORK="ipv6"

# Libvirt network for dual stack tests
export VM_DUAL_STACK_NETWORK="dual-stack"

# Location of RPMs built from source
# shellcheck disable=SC2034  # used elsewhere
RPM_SOURCE="${OUTPUTDIR}/rpmbuild"

# Location of RPMs built from source
# shellcheck disable=SC2034  # used elsewhere
NEXT_RPM_SOURCE="${OUTPUTDIR}/rpmbuild-fake-next-minor"

# Location of RPMs built from source
# shellcheck disable=SC2034  # used elsewhere
BASE_RPM_SOURCE="${OUTPUTDIR}/rpmbuild-base"

# Location of RPM packages downloaded from brew
# shellcheck disable=SC2034  # used elsewhere
BREW_RPM_SOURCE="${IMAGEDIR}/brew-rpms"

# Location of local repository used by composer
export LOCAL_REPO="${IMAGEDIR}/rpm-repos/microshift-local"

# Location of local repository used by composer
export NEXT_REPO="${IMAGEDIR}/rpm-repos/microshift-fake-next-minor"

# Location of local repository used by composer
export BASE_REPO="${IMAGEDIR}/rpm-repos/microshift-base"

# Location of local repository used by composer
export BREW_REPO="${IMAGEDIR}/rpm-repos/microshift-brew"

# Location of container images list for all the built images
export CONTAINER_LIST="${IMAGEDIR}/container-images-list"

# Location of the local mirror registry data
export MIRROR_REGISTRY_DIR="${IMAGEDIR}/mirror-registry"

# Location of container images in oci-dir format for all the bootc images
export BOOTC_IMAGE_DIR="${IMAGEDIR}/bootc-images"

# Location of images produced by bootc ISO build procedure
export BOOTC_ISO_DIR="${IMAGEDIR}/bootc-iso-images"

# Location of static delta images for bootc
export BOOTC_STATIC_DELTA="${IMAGEDIR}/bootc-static-delta"

# Location of data files created by the tools for managing scenarios
# as they are run.
#
# The CI system will override this, but we need a default for local
# use. Use the image directory, since that is already served by a web
# server.
#
# shellcheck disable=SC2034  # used elsewhere
SCENARIO_INFO_DIR="${SCENARIO_INFO_DIR:-${IMAGEDIR}/scenario-info}"

# Exclude CNCF Conformance tests from execution. These tests run
# in serial mode and they need a significant amount of time to
# complete. Setting this variable to true will exclude them from
# the scenario test list.
EXCLUDE_CNCF_CONFORMANCE="${EXCLUDE_CNCF_CONFORMANCE:-false}"

# The location of the robot framework virtualenv.
# The CI system will override this.
# shellcheck disable=SC2034  # used elsewhere
RF_VENV=${RF_VENV:-${OUTPUTDIR}/robotenv}

# The location of the gomplate binary.
export GOMPLATE=${OUTPUTDIR}/bin/gomplate

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

    ip=$(ip -f inet addr show "${bridge}" | grep inet)
    if [ -z "${ip}" ]; then
      ip=$(ip -f inet6 addr show "${bridge}" | grep global)
    fi
    # $ ip -f inet addr show virbr0
    # 10: virbr0: <NO-CARRIER,BROADCAST,MULTICAST,UP> mtu 1500 qdisc noqueue state DOWN group default qlen 1000
    #     inet 192.168.122.1/24 brd 192.168.122.255 scope global virbr0
    #        valid_lft forever preferred_lft forever
    # $ ip -f inet6 addr show virbr1
    # 14: virbr1: <NO-CARRIER,BROADCAST,MULTICAST,UP> mtu 1500 qdisc noqueue state DOWN group default qlen 1000
    #     inet6 2001:db8:dead:beef:fe::1/96 scope global
    #        valid_lft forever preferred_lft forever
    #     inet6 fe80::5054:ff:fe02:7e12/64 scope link proto kernel_ll
    #        valid_lft forever preferred_lft forever
    echo "${ip}" | awk '{print $2}' | cut -d/ -f1
}

# The IP address of the current host on the bridge used for the
# default network for libvirt VMs.
VM_BRIDGE_IP="$(get_vm_bridge_ip "default")"

# Web server port number
WEB_SERVER_PORT=8080

# Web server URL using VM bridge IP with fallback to host name
# shellcheck disable=SC2034  # used elsewhere
WEB_SERVER_URL="http://${VM_BRIDGE_IP:-$(hostname)}:${WEB_SERVER_PORT}"

# Mirror registry port number
export MIRROR_REGISTRY_PORT=5000

# Mirror registry URL using VM bridge IP with fallback to host name
MIRROR_REGISTRY_URL="${VM_BRIDGE_IP:-$(hostname)}:${MIRROR_REGISTRY_PORT}/microshift"
export MIRROR_REGISTRY_URL

get_build_branch() {
    local -r ocp_ver="$(grep ^OCP_VERSION "${ROOTDIR}/Makefile.version.$(uname -m).var"  | awk '{print $NF}' | awk -F. '{print $1"."$2}')"
    local -r cur_branch="$(git branch --show-current 2>/dev/null)"

    # Check if the current branch is derived from "main"
    local -r main_top=$(git rev-parse main 2>/dev/null)
    local -r main_base="$(git merge-base "${cur_branch}" main 2>/dev/null)"
    if [ "${main_top}" = "${main_base}" ] ; then
        echo "main"
        return
    fi

    # Check if the current branch is derived from "release-${ocp-ver}"
    local -r rel_top=$(git rev-parse "release-${ocp_ver}" 2>/dev/null)
    local -r rel_base="$(git merge-base "${cur_branch}" "release-${ocp_ver}" 2>/dev/null)"
    if [ "${rel_top}" = "${rel_base}" ] ; then
        echo "release-${ocp_ver}"
        return
    fi

    # Fallback to main if none of the above works
    echo "main"
}

# The branch identifier of the current scenario repository,
# i.e. "main", "release-4.14", etc.
# Used for top-level directory names when caching build artifacts,
# i.e. <bucket_name>/<branch>
# shellcheck disable=SC2034  # used elsewhere
SCENARIO_BUILD_BRANCH="$(get_build_branch)"

# The tag identifier of a scenario used in directory
# names when caching today's build artifacts,
# i.e. <bucket_name>/<branch>/<tag>
# shellcheck disable=SC2034  # used elsewhere
SCENARIO_BUILD_TAG="$(date '+%y%m%d')"

# The tag identifier of a scenario used in directory
# names when caching build artifacts from a day before,
# i.e. <bucket_name>/<branch>/<tag>
# shellcheck disable=SC2034  # used elsewhere
SCENARIO_BUILD_TAG_PREV="$(date -d "yesterday" '+%y%m%d')"

# The location of the awscli binary.
# shellcheck disable=SC2034  # used elsewhere
AWSCLI="${OUTPUTDIR}/bin/aws"

# The location of developer overrides files.
DEV_OVERRIDES="${TESTDIR}/dev_overrides.sh"
if [ -f "${DEV_OVERRIDES}" ]; then
    # The file will not exist, so we should not ask shellcheck to look for it.
    # shellcheck disable=SC1090
    source "${DEV_OVERRIDES}"
fi

# Return a scenario type based on its original path name
get_scenario_type_from_path() {
    local type

    case "${1}" in
    */scenarios/*|*/scenarios-ostree/*)
        type="ostree"
        ;;
    */scenarios-bootc/*)
        type="bootc"
        ;;
    */scenarios-bootc-containers/*)
        type="bootc-containers"
        ;;
    *)
        type="unknown"
        ;;
    esac
    echo "${type}"
}

# Prepare a static delta file between two container images specified in the
# function arguments. The images are copied from the mirror registry to a local
# directory and compared using the 'tar-diff' command.
#
# The result is stored in the '${BOOTC_STATIC_DELTA}/${src}@${dst}.tardiff' file.
#
# Arguments:
#   src - Source container image name
#   dst - Destination container image name
# Return:
#   '${src}@${dst}.tardiff' file stored in the '${BOOTC_STATIC_DELTA}' directory
prepare_static_delta() {
    local -r src=$1
    local -r dst=$2
    local -r tardiff="${src}@${dst}.tardiff"

    local -r workdir="$(mktemp -d /tmp/bootc-delta-prepare.XXXXXXXX)"
    # shellcheck disable=SC2164
    pushd "${workdir}" &>/dev/null

    "${ROOTDIR}/scripts/fetch_tools.sh" tar-diff

    # Export the images and archive them
    # Note: Cannot use 'docker-archive' because it only supports Docker Schema 2 manifests
    for i in "${src}" "${dst}" ; do
        skopeo copy --all --preserve-digests \
            --authfile "${MIRROR_REGISTRY_DIR}/config/microshift_auth.json" \
            docker://"${MIRROR_REGISTRY_URL}/${i}:latest" \
            oci-archive:"${i}.tar"
    done

    # Create a tar-diff file and list tar sizes
    "${ROOTDIR}/_output/bin/tar-diff" "${src}.tar" "${dst}.tar" "${tardiff}"
    ls -lh ./*.tar*

    # Store the tar-diff file in the static delta directory
    mkdir -p "${BOOTC_STATIC_DELTA}"
    mv -f "${tardiff}" "${BOOTC_STATIC_DELTA}/"

    # shellcheck disable=SC2164
    popd &>/dev/null
    rm -rf "${workdir}"
}

# Apply a static delta file created by the 'prepare_static_delta' function.
# The '${src}' image is copied from the mirror registry to a local directory
# and merged with the '${BOOTC_STATIC_DELTA}/${src}@${dst}.tardiff' file.
#
# The result of the merge is uploaded to the mirror registry and stored under
# the '${dst}-patched' name.
#
# Arguments:
#   src - Source container image name
#   dst - Destination container image name
# Return:
#   '${dst}-from-sdelta' image stored in the mirror registry
apply_static_delta() {
    local -r src=$1
    local -r dst=$2
    local -r tardiff="${src}@${dst}.tardiff"

    local -r workdir="$(mktemp -d /tmp/bootc-delta-apply.XXXXXXXX)"
    # shellcheck disable=SC2164
    pushd "${workdir}" &>/dev/null

    "${ROOTDIR}/scripts/fetch_tools.sh" tar-diff

    # Export the source image to be patched
    skopeo copy --all --preserve-digests \
        --authfile "${MIRROR_REGISTRY_DIR}/config/microshift_auth.json" \
        docker://"${MIRROR_REGISTRY_URL}/${src}:latest" \
        oci:"${src}"

    # Upgrade the source image to the destination using tar-patch and list tar sizes
    cp "${BOOTC_STATIC_DELTA}/${tardiff}" .
    "${ROOTDIR}/_output/bin/tar-patch" "${tardiff}" "${src}/" "${dst}-from-sdelta.tar"
    ls -lh ./*.tar*

    # Copy the patched image into the registry
    skopeo copy --all --preserve-digests \
        --authfile "${MIRROR_REGISTRY_DIR}/config/microshift_auth.json" \
        oci-archive:"${dst}-from-sdelta.tar" \
        docker://"${MIRROR_REGISTRY_URL}/${dst}-from-sdelta:latest"

    # shellcheck disable=SC2164
    popd &>/dev/null
    rm -rf "${workdir}"
}

# Lists of RPMs packages to build ostree and bootc images
MICROSHIFT_MANDATORY_RPMS_LIST=(
    microshift
    microshift-release-info
)
MICROSHIFT_Y2_OPTIONAL_RPMS_LIST=(
    microshift-olm
    microshift-olm-release-info
    microshift-multus
    microshift-multus-release-info
    microshift-gateway-api
    microshift-gateway-api-release-info
    microshift-low-latency
    microshift-observability
)
MICROSHIFT_Y1_OPTIONAL_RPMS_LIST=(
    "${MICROSHIFT_Y2_OPTIONAL_RPMS_LIST[@]}"
    microshift-cert-manager
    microshift-cert-manager-release-info
    microshift-sriov
    microshift-sriov-release-info
)
MICROSHIFT_OPTIONAL_RPMS_LIST=(
    "${MICROSHIFT_Y1_OPTIONAL_RPMS_LIST[@]}"
)
MICROSHIFT_Y2_X86_64_RPMS_LIST=(
    microshift-ai-model-serving
    microshift-ai-model-serving-release-info
)
MICROSHIFT_Y1_X86_64_RPMS_LIST=(
    "${MICROSHIFT_Y2_X86_64_RPMS_LIST[@]}"
)
MICROSHIFT_X86_64_RPMS_LIST=(
    "${MICROSHIFT_Y1_X86_64_RPMS_LIST[@]}"
)

export MICROSHIFT_MANDATORY_RPMS="${MICROSHIFT_MANDATORY_RPMS_LIST[*]}"
export MICROSHIFT_Y2_OPTIONAL_RPMS="${MICROSHIFT_Y2_OPTIONAL_RPMS_LIST[*]}"
export MICROSHIFT_Y1_OPTIONAL_RPMS="${MICROSHIFT_Y1_OPTIONAL_RPMS_LIST[*]}"
export MICROSHIFT_OPTIONAL_RPMS="${MICROSHIFT_OPTIONAL_RPMS_LIST[*]}"
export MICROSHIFT_Y1_X86_64_RPMS="${MICROSHIFT_Y1_X86_64_RPMS_LIST[*]}"
export MICROSHIFT_X86_64_RPMS="${MICROSHIFT_X86_64_RPMS_LIST[*]}"
