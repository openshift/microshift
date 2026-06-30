#!/bin/bash
set -eou pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOTDIR="$( cd "${SCRIPTDIR}/../../" && pwd )"
WORKTREE_BASE="${ROOTDIR}/.worktrees"

CONTAINERFILE="${SCRIPTDIR}/Containerfile.bootc-vm"

# Default pull secret location
PULL_SECRET="${PULL_SECRET:-${HOME}/.pull-secret.json}"

function resolve_names() {
    CURRENT_BRANCH=$(git -C "${ROOTDIR}" rev-parse --abbrev-ref HEAD)
    DEVENV_BRANCH="${DEVENV_BRANCH:-${CURRENT_BRANCH}}"

    # Fall back to main if branch is not in rhel-versions.json
    if ! jq -e --arg b "${DEVENV_BRANCH}" 'has($b)' "${SCRIPTDIR}/rhel-versions.json" &>/dev/null; then
        DEVENV_BRANCH="main"
    fi

    BRANCH_TAG="${DEVENV_BRANCH//\//-}"
    MACHINE_NAME="microshift-vm-${BRANCH_TAG}"
    IMAGE_NAME="microshift-vm:${BRANCH_TAG}"
}

function resolve_source() {
    resolve_names

    RHEL_VERSION=$(jq -r --arg b "${DEVENV_BRANCH}" '.[$b] // empty' "${SCRIPTDIR}/rhel-versions.json")
    if [ -z "${RHEL_VERSION}" ]; then
        echo "ERROR: No RHEL version for branch '${DEVENV_BRANCH}' in rhel-versions.json"
        exit 1
    fi

    SOURCE_DIR="${WORKTREE_BASE}/${BRANCH_TAG}"
    if [ ! -d "${SOURCE_DIR}" ]; then
        echo "Creating worktree for branch '${DEVENV_BRANCH}'..."
        mkdir -p "${WORKTREE_BASE}"
        git -C "${ROOTDIR}" worktree add --detach "${SOURCE_DIR}" "${DEVENV_BRANCH}"
    fi
}

function usage() {
    echo "Usage: $(basename "$0") <command>"
    echo ""
    echo "Commands:"
    echo "  setup    Build the VM image"
    echo "  start    Start the VM and configure it"
    echo "  stop     Stop the VM"
    echo "  delete   Delete the VM"
    echo "  shell    Open a shell in the VM"
    echo "  exec     Execute a command in the VM"
    echo "  status   Show the VM status"
    echo ""
    echo "Environment variables:"
    echo "  PULL_SECRET          Path to pull secret (default: ~/.pull-secret.json)"
    echo "  RHSM_ORG             Red Hat subscription org ID (required for 'setup' and 'start')"
    echo "  RHSM_ACTIVATION_KEY  Red Hat subscription activation key (required for 'setup' and 'start')"
    echo "  DEVENV_BRANCH        Build for a different branch (creates a worktree)"
    echo ""
    echo "Examples:"
    echo "  $(basename "$0") setup                                      # build using current branch"
    echo "  $(basename "$0") start                                      # start and configure the VM"
    echo "  DEVENV_BRANCH=release-4.21 $(basename "$0") setup           # build for release-4.21"
    echo "  DEVENV_BRANCH=release-4.21 $(basename "$0") shell           # shell into release-4.21 VM"

    [ -n "$1" ] && echo -e "\nERROR: $1"
    exit 1
}

function check_prerequisites() {
    if ! command -v podman &>/dev/null; then
        echo "ERROR: podman is not installed"
        exit 1
    fi
    if [ ! -f "${PULL_SECRET}" ]; then
        echo "ERROR: Pull secret not found at '${PULL_SECRET}'"
        echo "Set PULL_SECRET env var to the correct path"
        exit 1
    fi
    if [ ! -e /dev/kvm ]; then
        echo "ERROR: /dev/kvm not found — KVM is required for podman machine"
        exit 1
    fi
    if ! rpm -q qemu-kvm &>/dev/null; then
        sudo dnf install -y qemu-kvm
    fi
    mkdir -p "${HOME}/.local/bin"
    # RHEL installs these in /usr/libexec/ which is not on PATH
    for bin in "qemu-system-$(uname -m):/usr/libexec/qemu-kvm" \
               "virtiofsd:/usr/libexec/virtiofsd"; do
        local name="${bin%%:*}" target="${bin##*:}"
        if ! command -v "${name}" &>/dev/null && [ -x "${target}" ]; then
            ln -sfn "${target}" "${HOME}/.local/bin/${name}"
        fi
    done
}

function machine_exists() {
    podman machine inspect "${MACHINE_NAME}" &>/dev/null
}

function machine_running() {
    podman machine inspect "${MACHINE_NAME}" --format '{{.State}}' 2>/dev/null | grep -q "running"
}

function machine_exec() {
    podman machine ssh "${MACHINE_NAME}" -- "$@"
}

function cmd_setup() {
    check_prerequisites
    if [ -z "${RHSM_ORG:-}" ] || [ -z "${RHSM_ACTIVATION_KEY:-}" ]; then
        echo "ERROR: RHSM_ORG and RHSM_ACTIVATION_KEY env vars are required"
        exit 1
    fi

    local rhsm_org_file rhsm_key_file
    rhsm_org_file=$(mktemp)
    rhsm_key_file=$(mktemp)
    echo -n "${RHSM_ORG}" > "${rhsm_org_file}"
    echo -n "${RHSM_ACTIVATION_KEY}" > "${rhsm_key_file}"
    trap "rm -f '${rhsm_org_file}' '${rhsm_key_file}'" RETURN

    local -a build_args=(
        --authfile "${PULL_SECRET}"
        --build-arg "RHEL_VERSION=${RHEL_VERSION}"
        --secret "id=rhsm_org,src=${rhsm_org_file}"
        --secret "id=rhsm_key,src=${rhsm_key_file}"
        -t "${IMAGE_NAME}"
        -f "${CONTAINERFILE}"
    )

    echo "Building bootc image '${IMAGE_NAME}' (RHEL ${RHEL_VERSION}) for branch '${DEVENV_BRANCH}'..."
    sudo podman build "${build_args[@]}" "${ROOTDIR}"

    # Convert bootc image to qcow2 disk image using image-builder-cli
    DISK_IMAGE="${ROOTDIR}/_output/vm-images/${BRANCH_TAG}.qcow2"
    mkdir -p "$(dirname "${DISK_IMAGE}")"
    local output_dir
    output_dir=$(mktemp -d)
    trap "rm -rf '${output_dir}' '${rhsm_org_file}' '${rhsm_key_file}'" RETURN

    # Create blueprint with builder user and podman machine's SSH key
    local blueprint
    blueprint=$(mktemp --suffix=.toml)
    local ssh_pubkey
    ssh_pubkey=$(cat "${HOME}/.local/share/containers/podman/machine/machine.pub")
    cat > "${blueprint}" <<EOF
[[customizations.user]]
name = "builder"
key = "${ssh_pubkey}"
groups = ["wheel"]
EOF
    trap "rm -rf '${output_dir}' '${rhsm_org_file}' '${rhsm_key_file}' '${blueprint}'" RETURN

    echo "Building qcow2 disk image..."
    sudo podman run --rm -i --privileged \
        --network host \
        --security-opt label=type:unconfined_t \
        -v "${output_dir}:/output" \
        -v "${blueprint}:/blueprint.toml:ro" \
        -v /var/lib/containers/storage:/var/lib/containers/storage \
        ghcr.io/osbuild/image-builder-cli:latest \
        build --bootc-ref "localhost/${IMAGE_NAME}" \
            --blueprint /blueprint.toml \
            --output-dir /output \
            qcow2

    mv -f "${output_dir}"/*.qcow2 "${DISK_IMAGE}"
    echo "Disk image created at '${DISK_IMAGE}'"
}

function cmd_start() {
    check_prerequisites
    if [ -z "${RHSM_ORG:-}" ] || [ -z "${RHSM_ACTIVATION_KEY:-}" ]; then
        echo "ERROR: RHSM_ORG and RHSM_ACTIVATION_KEY env vars are required"
        exit 1
    fi

    # Already running
    if machine_running; then
        echo "VM '${MACHINE_NAME}' is already running"
        return 0
    fi

    # Exists but stopped — just start it
    if machine_exists; then
        echo "Starting stopped VM '${MACHINE_NAME}'..."
        podman machine start "${MACHINE_NAME}"
        echo "VM '${MACHINE_NAME}' started"
        return 0
    fi

    # Create VM from the qcow2 disk image
    DISK_IMAGE="${ROOTDIR}/_output/vm-images/${BRANCH_TAG}.qcow2"
    if [ ! -f "${DISK_IMAGE}" ]; then
        echo "ERROR: Disk image not found at '${DISK_IMAGE}'. Run '$(basename "$0") setup' first."
        exit 1
    fi

    echo "Creating VM '${MACHINE_NAME}'..."
    podman machine init \
        --cpus 4 \
        --memory 8192 \
        --disk-size 200 \
        --image "${DISK_IMAGE}" \
        --volume "${SOURCE_DIR}:/opt/microshift" \
        --volume "${ROOTDIR}/.git:${ROOTDIR}/.git" \
        --username builder \
        --rootful \
        "${MACHINE_NAME}"

    echo "Starting VM '${MACHINE_NAME}'..."
    podman machine start "${MACHINE_NAME}"

    # Copy pull secret into the VM
    podman machine cp "${PULL_SECRET}" "${MACHINE_NAME}:/etc/crio/openshift-pull-secret"
    podman machine cp "${PULL_SECRET}" "${MACHINE_NAME}:/home/builder/.pull-secret.json"

    # Configure git safe directory
    machine_exec sudo git config --system --add safe.directory /opt/microshift

    # Register subscription
    echo "Registering subscription..."
    machine_exec sudo subscription-manager register \
        --org="${RHSM_ORG}" \
        --activationkey="${RHSM_ACTIVATION_KEY}"

    # Run the release branch's configure-vm.sh
    echo "Running configure-vm.sh..."
    machine_exec bash -x /opt/microshift/scripts/devenv-builder/configure-vm.sh \
        --no-build --skip-dnf-update /etc/crio/openshift-pull-secret

    # Configure composer
    echo "Running manage_composer_config.sh..."
    machine_exec bash -x /opt/microshift/test/bin/manage_composer_config.sh create

    # Configure hypervisor for VM management
    echo "Running manage-vm.sh config and manage_hypervisor_config.sh create..."
    machine_exec bash -x -c '/opt/microshift/scripts/devenv-builder/manage-vm.sh config && /opt/microshift/test/bin/manage_hypervisor_config.sh create'

    echo "VM '${MACHINE_NAME}' started"
    echo "Use '$(basename "$0") shell' to open a shell"
}

function cmd_stop() {
    if ! machine_running; then
        echo "VM '${MACHINE_NAME}' is not running"
        return 0
    fi
    podman machine stop "${MACHINE_NAME}"
    echo "VM '${MACHINE_NAME}' stopped"
}

function cmd_delete() {
    if machine_running; then
        echo "ERROR: VM '${MACHINE_NAME}' is running. Run '$(basename "$0") stop' first."
        exit 1
    fi
    if ! machine_exists; then
        echo "VM '${MACHINE_NAME}' does not exist"
        return 0
    fi
    podman machine rm -f "${MACHINE_NAME}"
    echo "VM '${MACHINE_NAME}' deleted"
}

function cmd_shell() {
    if ! machine_running; then
        echo "ERROR: VM '${MACHINE_NAME}' is not running. Run '$(basename "$0") start' first."
        exit 1
    fi
    podman machine ssh "${MACHINE_NAME}"
}

function cmd_exec() {
    if [ $# -eq 0 ]; then
        usage "No command specified for exec"
    fi
    if ! machine_running; then
        echo "ERROR: VM '${MACHINE_NAME}' is not running. Run '$(basename "$0") start' first."
        exit 1
    fi
    machine_exec "$@"
}

function cmd_status() {
    if machine_running; then
        echo "VM '${MACHINE_NAME}': running"
        machine_exec sudo subscription-manager status 2>/dev/null \
            && echo "Subscription: registered" \
            || echo "Subscription: not registered"
    elif machine_exists; then
        echo "VM '${MACHINE_NAME}': stopped"
    else
        echo "VM '${MACHINE_NAME}': not found"
    fi
}

#
# MAIN
#
if [ $# -lt 1 ]; then
    usage "No command specified"
fi

while [ $# -gt 0 ]; do
    command="$1"
    shift

    case "${command}" in
        setup)  resolve_source; cmd_setup ;;
        start)  resolve_source; cmd_start ;;
        stop)   resolve_names; cmd_stop ;;
        delete) resolve_names; cmd_delete ;;
        shell)  resolve_names; cmd_shell ;;
        exec)   resolve_names; cmd_exec "$@"; break ;;
        status) resolve_names; cmd_status ;;
        *)      usage "Unknown command: ${command}" ;;
    esac
done
