#!/usr/bin/bash

set -eux

# --- Helper function for logging ---
log() {
    echo "[INFO] $(date +'%Y-%m-%d %H:%M:%S') - $1"
}

error() {
    echo "[ERROR] $(date +'%Y-%m-%d %H:%M:%S') - $1" >&2
}
# --- Helper function for logging ---

ssh_build_vm() {
    # shellcheck disable=SC2029
    # shellcheck disable=SC2086
    ssh -o USER=microshift -q "${VM_IPADDR}" "$@"
}

MICROSHIFT_REPO_OWNER=${1:-openshift}
MICROSHIFT_REPO_BRANCH="${2:-main}"
RPM_ORIGIN="${3:-all}"

CONFIG_VM=false
VM_NAME="microshift-dev"
REPO_NAME="microshift-${MICROSHIFT_REPO_BRANCH}-repo"

if [ ! -d "${HOME}/${REPO_NAME}" ]; then
    log "Cloning MicroShift repository..."
    git clone "https://github.com/${MICROSHIFT_REPO_OWNER}/microshift.git" --single-branch --branch "${MICROSHIFT_REPO_BRANCH}" "${HOME}/${REPO_NAME}"
else
    log "Directory ${HOME}/${REPO_NAME} already exists. Skipping clone."

fi

# Creating VM
if sudo virsh list | grep -q "${VM_NAME}"; then
    log "VM '${VM_NAME}' already exists. Skipping VM creation."
else
    log "Creating VM..."
    ./"${HOME}/${REPO_NAME}"/scripts/devenv-builder/manage-vm.sh config
    ./"${HOME}/${REPO_NAME}"/scripts/devenv-builder/manage-vm.sh create -n "${VM_NAME}"
fi

# Configuring VM
VM_IPADDR=$("${HOME}/${REPO_NAME}/scripts/devenv-builder/manage-vm.sh" ip -n "${VM_NAME}")
if ${CONFIG_VM}; then
    log "VM '${VM_NAME}' local IP address: ${VM_IPADDR}"
    ssh-copy-id -f "microshift@${VM_IPADDR}"
    scp "${HOME}/${REPO_NAME}/scripts/devenv-builder/configure-vm.sh" "microshift@${VM_IPADDR}:~/configure-vm.sh"
    scp "${HOME}/.pull-secret-microshift-dev.json" "microshift@${VM_IPADDR}:~/.pull-secret.json"
    ssh_build_vm "${HOME}/configure-vm.sh --no-build ${HOME}/.pull-secret.json"
fi

# Build rpms
ssh_build_vm "sudo rm -rf ${HOME}/.cache ${HOME}/.local"
if ( ! ssh_build_vm "test -d ${HOME}/${REPO_NAME}" ); then
    log "Cloning MicroShift repository..."
    ssh_build_vm git clone "https://github.com/${MICROSHIFT_REPO_OWNER}/microshift.git" --single-branch --branch "${MICROSHIFT_REPO_BRANCH}" "${HOME}/${REPO_NAME}"
else
    log "Directory ${HOME}/${REPO_NAME} already exists. Skipping clone."
    ssh_build_vm git -C "${HOME}/${REPO_NAME}" fetch
    ssh_build_vm git -C "${HOME}/${REPO_NAME}" checkout "${MICROSHIFT_REPO_BRANCH}"
    ssh_build_vm git -C "${HOME}/${REPO_NAME}" pull
fi

if ( ! ssh_build_vm "test -d ${HOME}/${REPO_NAME}/_output/test-images/rpm-repos/" ); then
    ssh_build_vm "bash -x ${HOME}/${REPO_NAME}/test/bin/build_rpms.sh ${RPM_ORIGIN}"
    log "RPM packages available: ${HOME}/${REPO_NAME}/_output/test-images/rpm-repos/"
else
    log "RPM packages available: ${HOME}/${REPO_NAME}/_output/test-images/rpm-repos/"
    ssh_build_vm "ls -lrth ${HOME}/${REPO_NAME}/_output/test-images/rpm-repos/"
fi

# Copy RPM packages into QE microshift build versions
MICROSHIFT_VERSION=$(ssh_build_vm ls "${HOME}/${REPO_NAME}/_output/test-images/rpm-repos/microshift-brew/microshift-4*" | sed -n 's/.*microshift-\(.*\)-.*/\1/p')
log "MICROSHIFT_VERSION=${MICROSHIFT_VERSION}"
mkdir -p "${HOME}/libvirt/images/microshift_builds/${MICROSHIFT_VERSION}/"
scp -qr "microshift@${VM_IPADDR}:${HOME}/${REPO_NAME}/_output/test-images/rpm-repos/" "${HOME}/libvirt/images/microshift_builds/${MICROSHIFT_VERSION}/rpm-repos/"


