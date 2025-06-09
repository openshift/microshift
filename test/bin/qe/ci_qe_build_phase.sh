#!/usr/bin/bash

########### usage ###########
#
# This script MUST be executed on the hypervisor, it will run these steps:
#  1) create build VM
#  2) configure build VM
#  3) git clone/fetch microshift repo into the build VM
#  4) build rpm packages
#  5) build bootc images
#  6) copy RPM packages out from build VM into the hypervisor
#
# Example: ./ci_qe_build_phase.sh agullon USHIFT-5397 brew 4.19.0~rc.4
#
#############################

set -euo pipefail

########### Functions section ###########

function log() {
    echo "[INFO] $(date +'%Y-%m-%d %H:%M:%S') - $1"
}

function error() {
    echo "[ERROR] $(date +'%Y-%m-%d %H:%M:%S') - $1" >&2
}

function ssh_build_vm() {
    # shellcheck disable=SC2029
    # shellcheck disable=SC2086
    ssh -o USER=microshift -q "${VM_IPADDR}" "$@"
}

function create_vm() {
    if sudo virsh list | grep -q "${VM_NAME}"; then
        log "VM '${VM_NAME}' already exists. Skipping VM creation."
    else
        log "Creating VM..."
        ${EXE_MODE} "${SCRIPTDIR}"/../../../scripts/devenv-builder/manage-vm.sh config
        ${EXE_MODE} "${SCRIPTDIR}"/../../../scripts/devenv-builder/manage-vm.sh create -n "${VM_NAME}"
    fi
    VM_IPADDR=$("${SCRIPTDIR}/../../../scripts/devenv-builder/manage-vm.sh" ip -n "${VM_NAME}")
    log "VM '${VM_NAME}' IP address: ${VM_IPADDR}"
}

function configure_vm() {
    if ${CONFIG_VM_FLAG}; then
        ssh-copy-id -f "microshift@${VM_IPADDR}"
        scp "${REPO_PATH}/scripts/devenv-builder/configure-vm.sh" "microshift@${VM_IPADDR}:~/configure-vm.sh"
        scp "${HOME}/.pull-secret-microshift-dev.json" "microshift@${VM_IPADDR}:~/.pull-secret.json"
        ssh_build_vm "${HOME}/configure-vm.sh --no-build ${HOME}/.pull-secret.json"
    fi
}

function fetch_microshift() {
    LOCAL_REPO_PATH="$(realpath ${SCRIPTDIR}/../../../)"

    ssh_build_vm "sudo rm -rf ${HOME}/microshift"
    ssh_build_vm "sudo rm -rf ${HOME}/.cache ${HOME}/.local"
    ssh_build_vm "sudo rm -rf /var/cache /var/log"
    
    scp -qr "${LOCAL_REPO_PATH}" "microshift@${VM_IPADDR}:${REPO_PATH}/"
    
    # if ( ! ssh_build_vm "test -d ${REPO_PATH}" ); then
    #     log "Cloning MicroShift repository inside the build VM..."
    #     ssh_build_vm git clone "https://github.com/${MICROSHIFT_REPO_OWNER}/microshift.git" --depth 1 --single-branch --branch "${MICROSHIFT_REPO_BRANCH}" "${REPO_PATH}"
    # else
    #     log "Directory ${REPO_PATH} already exists in build VM. Skipping clone."
    #     ssh_build_vm git -C "${REPO_PATH}" fetch
    #     ssh_build_vm git -C "${REPO_PATH}" checkout "${MICROSHIFT_REPO_BRANCH}"
    #     ssh_build_vm git -C "${REPO_PATH}" pull
    # fi
}

function build_rpms() {
    ssh_build_vm "${EXE_MODE} ${REPO_PATH}/test/bin/build_rpms.sh ${RPM_SOURCE} ${MICROSHIFT_BREW_VERSION}"
    log "RPM packages available: ${REPO_PATH}/_output/test-images/rpm-repos/"
}

function build_bootc_images() {
    if ${BUILD_BOOTC_IMAGES_FLAG}; then
        ssh_build_vm "make -C "${REPO_PATH}" verify-containers"

        ssh_build_vm "${EXE_MODE} ${REPO_PATH}/bin/build_bootc_images.sh -l ./image-blueprints-bootc/layer1-base"
        # ssh_build_vm "${EXE_MODE} ${REPO_PATH}/bin/build_bootc_images.sh -l ./image-blueprints-bootc/layer2-presubmit"
        # ssh_build_vm "${EXE_MODE} ${REPO_PATH}/bin/build_bootc_images.sh -l ./image-blueprints-bootc/layer3-periodic"

        log "BootC images available: ${REPO_PATH}/_output/test-images/______"
    else
        log "BootC images available: ${REPO_PATH}/_output/test-images/______"
        # ssh_build_vm "ls -lrth ${REPO_PATH}/_output/test-images/______"
fi
}

function copy_artifacts() {
    MICROSHIFT_VERSION=$(ssh_build_vm ls "${REPO_PATH}/_output/test-images/rpm-repos/microshift-brew/microshift-4*" | sed -n 's/.*microshift-\(.*\)-.*/\1/p')
    log "MICROSHIFT_VERSION=${MICROSHIFT_VERSION}"
    mkdir -p "${HOME}/libvirt/images/microshift_builds/${MICROSHIFT_VERSION}/"
    scp -qr "microshift@${VM_IPADDR}:${REPO_PATH}/_output/test-images/rpm-repos/" "${HOME}/libvirt/images/microshift_builds/${MICROSHIFT_VERSION}/"
    log "Artifacts copied into the hypervisor: ${HOME}/libvirt/images/microshift_builds/${MICROSHIFT_VERSION}/"
}


# default params
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
VM_NAME="microshif-build-vm_${TIMESTAMP}"
SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPO_PATH="microshift"
EXE_MODE="bash"

# input param
MICROSHIFT_REPO_OWNER="openshift"
MICROSHIFT_REPO_BRANCH="main"
RPM_SOURCE="brew"
MICROSHIFT_BREW_VERSION="" # By default it will fetch last brew available from this version

# input flags
CONFIG_VM_FLAG=false
BUILD_BOOTC_IMAGES_FLAG=false

function usage() {
    echo "Usage: $(basename "$0") [--repo-owner <owner>] [--repo_branch <branch>] [--rpm-source <brew|source|all>] [--microshift-brew-version <version>] [--vm-name <vm_name>]"
    echo ""
    echo "  --repo-owner                  MicroShift repository owner. Default: openshift"
    echo "  --repo-branch                 MicroShift repository branch. Default: main"
    echo "  --rpm-source                  RPMs packages source. Default: brew. Valid values: brew, source and all"
    echo "  --microshift-brew-version     MicroShift version to fetch from brew. Default is empty, which means latest. Example: 4.19"
    echo "  --vm-name                     Name of the build VM. Default value: "microshif-build-vm_${TIMESTAMP}""
    echo "  --verbose                     Run bash scripts with -x flag."

    [ -n "$1" ] && echo -e "\nERROR: $1"
    exit 1
}

while [ $# -gt 0 ]; do
    case "$1" in
    --repo-owner)
        MICROSHIFT_REPO_OWNER=$2
        shift 2
        ;;
    --repo-branch)
        MICROSHIFT_REPO_BRANCH=$2
        shift 2
        ;;
    --rpm-source)
        RPM_SOURCE=$2
        shift 2
        ;;
    --microshift-brew-version)
        MICROSHIFT_BREW_VERSION=$2
        shift 2
        ;;
    --vm-name)
        VM_NAME=$2
        shift 2
        ;;
    --verbose)
        EXE_MODE="bash -x"
        set -x
        shift 1
        ;;
    *) usage ;;
    esac
done

if [ $# -ne 0 ]; then
    usage "Wrong number of arguments"
fi

########### Call functions section ###########
create_vm
configure_vm
fetch_microshift
build_rpms
build_bootc_images
copy_artifacts
