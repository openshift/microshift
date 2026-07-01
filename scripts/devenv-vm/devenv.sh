#!/bin/bash
set -eou pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOTDIR="$( cd "${SCRIPTDIR}/../../" && pwd )"

CONTAINERFILE="${SCRIPTDIR}/Containerfile.vm"

# Default pull secret location
PULL_SECRET="${PULL_SECRET:-${HOME}/.pull-secret.json}"

# VM resources
VM_CPUS="${VM_CPUS:-4}"
VM_MEMORY="${VM_MEMORY:-8192}"

function resolve_names() {
    CURRENT_BRANCH=$(git -C "${ROOTDIR}" rev-parse --abbrev-ref HEAD)
    DEVENV_BRANCH="${DEVENV_BRANCH:-${CURRENT_BRANCH}}"

    # Fall back to main if branch is not in rhel-versions.json
    if ! jq -e --arg b "${DEVENV_BRANCH}" 'has($b)' "${SCRIPTDIR}/rhel-versions.json" &>/dev/null; then
        DEVENV_BRANCH="main"
    fi

    BRANCH_TAG="${DEVENV_BRANCH//\//-}"
    RHEL_VERSION=$(jq -r --arg b "${DEVENV_BRANCH}" '.[$b] // empty' "${SCRIPTDIR}/rhel-versions.json")
    IMAGE_NAME="microshift-vm:${BRANCH_TAG}"
    VM_NAME="microshift-devenv-${BRANCH_TAG}"
    VM_DIR="${ROOTDIR}/_output/${VM_NAME}"
}

function usage() {
    echo "Usage: $(basename "$0") <command>"
    echo ""
    echo "Commands:"
    echo "  setup    Build the VM disk image"
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
    echo "  VM_CPUS              Number of CPUs (default: 4)"
    echo "  VM_MEMORY            Memory in MiB (default: 8192)"
    echo ""
    echo "Examples:"
    echo "  $(basename "$0") setup                                      # build using current branch"
    echo "  $(basename "$0") start                                      # start and configure the VM"
    echo "  $(basename "$0") shell                                      # shell into the VM"
    echo "  DEVENV_BRANCH=release-4.21 $(basename "$0") setup           # build for release-4.21"

    [ -n "$1" ] && echo -e "\nERROR: $1"
    exit 1
}

function resolve_rhsm() {
    if [ -z "${RHSM_ORG:-}" ] || [ -z "${RHSM_ACTIVATION_KEY:-}" ]; then
        echo "ERROR: RHSM_ORG and RHSM_ACTIVATION_KEY env vars are required"
        exit 1
    fi
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
        echo "ERROR: /dev/kvm not found — KVM is required"
        exit 1
    fi
    if ! command -v virt-install &>/dev/null || ! command -v virsh &>/dev/null; then
        echo "Installing libvirt prerequisites..."
        "${ROOTDIR}/scripts/devenv-builder/manage-vm.sh" config
    fi
    if ! sudo systemctl is-active --quiet nfs-server; then
        echo "Starting nfs-server..."
        sudo dnf install -y nfs-utils 2>/dev/null
        sudo systemctl enable --now nfs-server
    fi
    if command -v firewall-cmd &>/dev/null; then
        for svc in nfs mountd rpc-bind; do
            if ! sudo firewall-cmd --zone=libvirt --query-service="${svc}" --quiet 2>/dev/null; then
                sudo firewall-cmd --zone=libvirt --add-service="${svc}" --permanent --quiet
            fi
        done
        sudo firewall-cmd --reload --quiet
    fi
}

function vm_exists() {
    sudo virsh dominfo "${VM_NAME}" &>/dev/null
}

function vm_running() {
    sudo virsh domstate "${VM_NAME}" 2>/dev/null | grep -q "running"
}

function vm_ip() {
    if [ -f "${VM_DIR}/vm_ip" ]; then
        cat "${VM_DIR}/vm_ip"
        return
    fi
    local ip
    for i in $(seq 30); do
        ip=$(sudo virsh domifaddr "${VM_NAME}" 2>/dev/null | awk '/ipv4/ {split($4,a,"/"); print a[1]}' | head -1)
        if [ -n "${ip}" ]; then
            echo "${ip}" > "${VM_DIR}/vm_ip"
            echo "${ip}"
            return
        fi
        sleep 2
    done
    echo "ERROR: Could not determine VM IP address" >&2
    return 1
}

function vm_ssh() {
    local ip
    ip=$(vm_ip) || exit 1
    ssh -i "${VM_DIR}/ssh_key" \
        -o StrictHostKeyChecking=no \
        -o UserKnownHostsFile=/dev/null \
        -o LogLevel=ERROR \
        "builder@${ip}" "$@"
}

function vm_exec() {
    vm_ssh -- "$@"
}

function vm_exec_as_microshift() {
    local ip
    ip=$(vm_ip) || exit 1
    ssh -i "${VM_DIR}/ssh_key" \
        -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o LogLevel=ERROR \
        "microshift@${ip}" "cd /var/microshift && $(printf '%q ' "$@")"
}

function wait_for_ssh() {
    local ip="$1"
    echo "Waiting for SSH to become available on ${ip}..."
    for i in $(seq 60); do
        if ssh -i "${VM_DIR}/ssh_key" \
            -o StrictHostKeyChecking=no \
            -o UserKnownHostsFile=/dev/null \
            -o LogLevel=ERROR \
            -o ConnectTimeout=2 \
            "builder@${ip}" true 2>/dev/null; then
            echo "SSH is ready"
            return 0
        fi
        sleep 2
    done
    echo "ERROR: Timed out waiting for SSH"
    exit 1
}

function cmd_setup() {
    check_prerequisites

    mkdir -p "${VM_DIR}"

    # Generate SSH keypair if not present
    if [ ! -f "${VM_DIR}/ssh_key" ]; then
        echo "Generating SSH keypair..."
        ssh-keygen -t ed25519 -f "${VM_DIR}/ssh_key" -N "" -q
    fi

    # Generate random password for builder user (console access)
    if [ ! -f "${VM_DIR}/builder_password" ]; then
        openssl rand -base64 12 > "${VM_DIR}/builder_password"
        chmod 600 "${VM_DIR}/builder_password"
    fi

    # Create worktree for the build so configure scripts come from the target branch
    local worktree_base="${ROOTDIR}/.worktrees"
    local build_dir="${worktree_base}/${VM_NAME}"
    if [ ! -d "${build_dir}" ]; then
        echo "Creating worktree for branch '${DEVENV_BRANCH}'..."
        mkdir -p "${worktree_base}"
        git -C "${ROOTDIR}" worktree add --detach "${build_dir}" "${DEVENV_BRANCH}"
    fi

    # Build the bootc container image
    local rhsm_org_file rhsm_key_file
    rhsm_org_file=$(mktemp)
    rhsm_key_file=$(mktemp)
    echo -n "${RHSM_ORG}" > "${rhsm_org_file}"
    echo -n "${RHSM_ACTIVATION_KEY}" > "${rhsm_key_file}"
    trap "rm -f '${rhsm_org_file}' '${rhsm_key_file}'" EXIT

    local -a build_args=(
        --authfile "${PULL_SECRET}"
        --build-arg "RHEL_VERSION=${RHEL_VERSION}"
        --secret "id=rhsm_org,src=${rhsm_org_file}"
        --secret "id=rhsm_key,src=${rhsm_key_file}"
        --secret "id=pull_secret,src=${PULL_SECRET}"
        --secret "id=builder_password,src=${VM_DIR}/builder_password"
        --secret "id=ssh_pub_key,src=${VM_DIR}/ssh_key.pub"
        -t "${IMAGE_NAME}"
        -f "${CONTAINERFILE}"
    )

    echo "Building bootc image '${IMAGE_NAME}' (RHEL ${RHEL_VERSION}) from '${build_dir}'..."
    sudo podman build "${build_args[@]}" "${build_dir}"

    # Create blueprint with builder user and SSH key
    local blueprint
    blueprint=$(mktemp --suffix=.toml)
    cat > "${blueprint}" <<EOF
[customizations.kernel]
append = "systemd.zram=0"
EOF

    # Convert bootc image to qcow2 using image-builder-cli
    local output_dir
    output_dir=$(mktemp -d)

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

    # Move the qcow2 to the VM directory as the base image and fix ownership
    sudo mv -f "${output_dir}"/*.qcow2 "${VM_DIR}/base.qcow2"
    sudo chown "$(id -u):$(id -g)" "${VM_DIR}/base.qcow2"
    rm -rf "${output_dir}" "${blueprint}"

    echo "Base image created at '${VM_DIR}/base.qcow2'"
}

function cmd_start() {
    check_prerequisites

    # Already running
    if vm_running; then
        echo "VM '${VM_NAME}' is already running"
        return 0
    fi

    # Exists but stopped — start and remount NFS
    if vm_exists; then
        echo "Starting stopped VM '${VM_NAME}'..."
        sudo virsh start "${VM_NAME}"
        rm -f "${VM_DIR}/vm_ip"
        local ip
        ip=$(vm_ip)
        wait_for_ssh "${ip}"
        local host_ip
        host_ip=$(sudo virsh net-dumpxml default | grep -oP "ip address='\K[^']+")
        vm_exec sudo mkdir -p /var/microshift
        vm_exec sudo mount -t nfs "${host_ip}:${ROOTDIR}" /var/microshift
        echo "VM '${VM_NAME}' started"
        return 0
    fi

    if [ ! -f "${VM_DIR}/base.qcow2" ]; then
        echo "ERROR: Base image not found. Run '$(basename "$0") setup' first."
        exit 1
    fi

    # Create a working copy of the base image and resize for dev use
    echo "Creating VM disk from base image..."
    cp -f "${VM_DIR}/base.qcow2" "${VM_DIR}/disk.qcow2"
    qemu-img resize "${VM_DIR}/disk.qcow2" 50G

    # Export the project root via NFS
    if ! grep -Fqs "${ROOTDIR}" /etc/exports; then
        echo "${ROOTDIR} *(rw,sync,no_subtree_check,no_root_squash,insecure)" | sudo tee -a /etc/exports > /dev/null
        sudo exportfs -ra
    fi

    echo "Creating and starting VM '${VM_NAME}'..."
    sudo virt-install \
        --name "${VM_NAME}" \
        --vcpus "${VM_CPUS}" \
        --memory "${VM_MEMORY}" \
        --disk "path=${VM_DIR}/disk.qcow2,format=qcow2" \
        --import \
        --os-variant "rhel9-unknown" \
        --network network=default \
        --noautoconsole

    # Get VM IP and wait for SSH
    local ip
    echo "Waiting for VM IP..."
    ip=$(vm_ip)
    wait_for_ssh "${ip}"

    # Get the host IP on the libvirt network
    local host_ip
    host_ip=$(sudo virsh net-dumpxml default | grep -oP "ip address='\K[^']+")

    # Mount the project root via NFS
    vm_exec sudo mkdir -p /var/microshift
    vm_exec sudo mount -t nfs "${host_ip}:${ROOTDIR}" /var/microshift

    # Create microshift user matching host UID/GID with /var/microshift as home
    local host_uid host_gid
    host_uid=$(id -u)
    host_gid=$(id -g)
    vm_exec sudo groupadd -g "${host_gid}" microshift 2>/dev/null || true
    vm_exec sudo useradd -m -u "${host_uid}" -g "${host_gid}" -d /var/home/microshift -s /bin/bash -G wheel microshift 2>/dev/null || true
    vm_ssh -- "echo 'microshift ALL=(ALL) NOPASSWD: ALL' | sudo tee /etc/sudoers.d/microshift > /dev/null"
    vm_exec sudo cp /etc/ssh/authorized_keys/builder /etc/ssh/authorized_keys/microshift
    vm_exec sudo git config --system --add safe.directory '*'

    # Copy pull secret
    scp -i "${VM_DIR}/ssh_key" \
        -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o LogLevel=ERROR \
        "${PULL_SECRET}" "builder@${ip}:/tmp/pull-secret.json"
    vm_exec sudo mkdir -p /etc/crio
    vm_exec sudo cp /tmp/pull-secret.json /etc/crio/openshift-pull-secret
    vm_exec sudo cp /tmp/pull-secret.json /var/home/microshift/.pull-secret.json
    vm_exec sudo chown "${host_uid}:${host_gid}" /var/home/microshift/.pull-secret.json
    vm_exec rm -f /tmp/pull-secret.json

    # Register subscription
    echo "Registering subscription..."
    vm_exec sudo subscription-manager register --force \
        --org="${RHSM_ORG}" \
        --activationkey="${RHSM_ACTIVATION_KEY}"

    echo "VM '${VM_NAME}' started"
    echo "Use '$(basename "$0") shell' to open a shell"
}

function cmd_stop() {
    if ! vm_exists; then
        echo "VM '${VM_NAME}' does not exist"
        return 0
    fi
    if ! vm_running; then
        echo "VM '${VM_NAME}' is not running"
        return 0
    fi
    sudo virsh shutdown "${VM_NAME}"
    echo "Waiting for VM to shut down..."
    for i in $(seq 30); do
        if ! vm_running; then
            echo "VM '${VM_NAME}' stopped"
            rm -f "${VM_DIR}/vm_ip"
            return 0
        fi
        sleep 2
    done
    echo "Graceful shutdown timed out, forcing..."
    sudo virsh destroy "${VM_NAME}"
    rm -f "${VM_DIR}/vm_ip"
    echo "VM '${VM_NAME}' stopped"
}

function cmd_delete() {
    if vm_running; then
        echo "ERROR: VM '${VM_NAME}' is running. Run '$(basename "$0") stop' first."
        exit 1
    fi
    if vm_exists; then
        sudo virsh undefine "${VM_NAME}"
        rm -f "${VM_DIR}/vm_ip"
        echo "VM '${VM_NAME}' deleted (disk image preserved)"
    else
        echo "VM '${VM_NAME}' does not exist"
    fi
}

function cmd_shell() {
    if ! vm_running; then
        echo "ERROR: VM '${VM_NAME}' is not running. Run '$(basename "$0") start' first."
        exit 1
    fi
    local ip
    ip=$(vm_ip) || exit 1
    ssh -i "${VM_DIR}/ssh_key" \
        -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o LogLevel=ERROR \
        -t "microshift@${ip}" "cd /var/microshift && exec bash --login"
}

function cmd_exec() {
    if [ $# -eq 0 ]; then
        usage "No command specified for exec"
    fi
    if ! vm_running; then
        echo "ERROR: VM '${VM_NAME}' is not running. Run '$(basename "$0") start' first."
        exit 1
    fi
    vm_exec_as_microshift "$@"
}

function cmd_status() {
    if vm_running; then
        local ip
        ip=$(vm_ip 2>/dev/null) || ip="unknown"
        echo "VM '${VM_NAME}': running (IP ${ip})"
    elif vm_exists; then
        echo "VM '${VM_NAME}': stopped"
    elif [ -f "${VM_DIR}/base.qcow2" ]; then
        echo "VM '${VM_NAME}': not defined (base image exists)"
    else
        echo "VM '${VM_NAME}': not found"
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
        setup)  resolve_rhsm; resolve_names; cmd_setup ;;
        start)  resolve_rhsm; resolve_names; cmd_start ;;
        stop)   resolve_names; cmd_stop ;;
        delete) resolve_names; cmd_delete ;;
        shell)  resolve_names; cmd_shell ;;
        exec)   resolve_names; cmd_exec "$@"; break ;;
        status) resolve_names; cmd_status ;;
        *)      usage "Unknown command: ${command}" ;;
    esac
done
