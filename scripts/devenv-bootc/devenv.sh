#!/bin/bash
set -eou pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOTDIR="$( cd "${SCRIPTDIR}/../../" && pwd )"
WORKTREE_BASE="${ROOTDIR}/.worktrees"

CONTAINERFILE="${SCRIPTDIR}/Containerfile.bootc-builder"

# Default pull secret location
PULL_SECRET="${PULL_SECRET:-${HOME}/.pull-secret.json}"

# Common exec flags for builder user commands
CONTAINER_EXEC=(sudo podman exec --user builder --workdir /opt/microshift)

function resolve_names() {
    CURRENT_BRANCH=$(git -C "${ROOTDIR}" rev-parse --abbrev-ref HEAD)
    DEVENV_BRANCH="${DEVENV_BRANCH:-${CURRENT_BRANCH}}"

    # Fall back to main if branch is not in rhel-versions.json
    if ! jq -e --arg b "${DEVENV_BRANCH}" 'has($b)' "${SCRIPTDIR}/rhel-versions.json" &>/dev/null; then
        DEVENV_BRANCH="main"
    fi

    BRANCH_TAG="${DEVENV_BRANCH//\//-}"
    CONTAINER_NAME="microshift-builder-${BRANCH_TAG}"
    IMAGE_NAME="microshift-builder:${BRANCH_TAG}"
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
    echo "  setup    Build the builder container image"
    echo "  start    Start the builder container"
    echo "  stop     Stop the builder container"
    echo "  delete   Delete the builder container"
    echo "  shell    Open a shell in the builder container"
    echo "  exec     Execute a command in the builder container"
    echo "  status   Show the builder container status"
    echo ""
    echo "Environment variables:"
    echo "  PULL_SECRET          Path to pull secret (default: ~/.pull-secret.json)"
    echo "  RHSM_ORG             Red Hat subscription org ID (required for 'setup' and 'start')"
    echo "  RHSM_ACTIVATION_KEY  Red Hat subscription activation key (required for 'setup' and 'start')"
    echo "  DEVENV_BRANCH        Build for a different branch (creates a worktree)"
    echo ""
    echo "Examples:"
    echo "  $(basename "$0") setup                                      # build using current branch"
    echo "  $(basename "$0") start                                      # start and configure the container"
    echo "  DEVENV_BRANCH=release-4.21 $(basename "$0") setup           # build for release-4.21"
    echo "  DEVENV_BRANCH=release-4.21 $(basename "$0") shell           # shell into release-4.21 container"

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

    echo "Building builder image '${IMAGE_NAME}' (RHEL ${RHEL_VERSION}) for branch '${DEVENV_BRANCH}'..."
    sudo podman build "${build_args[@]}" "${ROOTDIR}"
}

function cmd_start() {
    check_prerequisites
    if [ -z "${RHSM_ORG:-}" ] || [ -z "${RHSM_ACTIVATION_KEY:-}" ]; then
        echo "ERROR: RHSM_ORG and RHSM_ACTIVATION_KEY env vars are required"
        exit 1
    fi

    # Reattach if already running
    if sudo podman ps --format '{{.Names}}' | grep -xFq "${CONTAINER_NAME}"; then
        echo "Container '${CONTAINER_NAME}' is already running"
        return 0
    fi

    # Restart a stopped container (preserves configured state)
    if sudo podman ps -a --format '{{.Names}}' | grep -xFq "${CONTAINER_NAME}"; then
        echo "Restarting stopped container '${CONTAINER_NAME}'..."
        sudo podman start "${CONTAINER_NAME}"
        echo "Waiting for systemd to boot..."
        sudo podman exec "${CONTAINER_NAME}" systemctl is-system-running --wait &>/dev/null || true
        echo "Container '${CONTAINER_NAME}' started"
        return 0
    fi

    local -a run_args=(
        -d
        --name "${CONTAINER_NAME}"
        --privileged
        --systemd=true
        --volume "${SOURCE_DIR}:/opt/microshift:z"
        --volume "${ROOTDIR}/.git:${ROOTDIR}/.git:z"
        --volume /var/tmp:/tmp
        --volume "${PULL_SECRET}:/etc/crio/openshift-pull-secret:z"
        --volume "${PULL_SECRET}:/home/builder/.pull-secret.json:z"
    )

    # Share AWS credentials if available
    if [ -d "${HOME}/.aws" ]; then
        run_args+=(--volume "${HOME}/.aws:/home/builder/.aws:z")
    fi

    echo "Starting builder container..."
    sudo podman run "${run_args[@]}" "${IMAGE_NAME}"

    # Match builder UID/GID to the host user so bind-mounted files have correct ownership
    local host_uid host_gid
    host_uid=$(id -u)
    host_gid=$(id -g)
    sudo podman exec "${CONTAINER_NAME}" usermod -u "${host_uid}" builder
    sudo podman exec "${CONTAINER_NAME}" groupmod -g "${host_gid}" builder

    # Symlink the host worktree path to /opt/microshift so git references resolve
    sudo podman exec "${CONTAINER_NAME}" \
        bash -c "mkdir -p '$(dirname "${SOURCE_DIR}")' && ln -sfn /opt/microshift '${SOURCE_DIR}'"
    sudo podman exec "${CONTAINER_NAME}" \
        git config --system --add safe.directory /opt/microshift

    # Wait for systemd to boot
    echo "Waiting for systemd to boot..."
    sudo podman exec "${CONTAINER_NAME}" systemctl is-system-running --wait &>/dev/null || true

    # Register subscription
    echo "Registering subscription..."
    "${CONTAINER_EXEC[@]}" "${CONTAINER_NAME}" \
        sudo subscription-manager register \
            --org="${RHSM_ORG}" \
            --activationkey="${RHSM_ACTIVATION_KEY}"

    # Run the release branch's configure-vm.sh
    echo "Running configure-vm.sh..."
    "${CONTAINER_EXEC[@]}" "${CONTAINER_NAME}" \
        bash -x /opt/microshift/scripts/devenv-builder/configure-vm.sh \
            --no-build --skip-dnf-update /etc/crio/openshift-pull-secret

    # Configure composer
    echo "Running manage_composer_config.sh..."
    "${CONTAINER_EXEC[@]}" "${CONTAINER_NAME}" \
        bash -x /opt/microshift/test/bin/manage_composer_config.sh create

    echo "Container '${CONTAINER_NAME}' started"
    echo "Use '$(basename "$0") shell' to open a shell"
}

function cmd_stop() {
    if ! sudo podman ps --format '{{.Names}}' | grep -xFq "${CONTAINER_NAME}"; then
        echo "Container '${CONTAINER_NAME}' is not running"
        return 0
    fi
    sudo podman stop "${CONTAINER_NAME}"
    echo "Container '${CONTAINER_NAME}' stopped"
}

function cmd_delete() {
    if sudo podman ps --format '{{.Names}}' | grep -xFq "${CONTAINER_NAME}"; then
        echo "ERROR: Container '${CONTAINER_NAME}' is running. Run '$(basename "$0") stop' first."
        exit 1
    fi
    if ! sudo podman ps -a --format '{{.Names}}' | grep -xFq "${CONTAINER_NAME}"; then
        echo "Container '${CONTAINER_NAME}' does not exist"
        return 0
    fi
    sudo podman rm -f "${CONTAINER_NAME}"
    echo "Container '${CONTAINER_NAME}' deleted"
}

function cmd_shell() {
    if ! sudo podman ps --format '{{.Names}}' | grep -xFq "${CONTAINER_NAME}"; then
        echo "ERROR: Container '${CONTAINER_NAME}' is not running. Run '$(basename "$0") start' first."
        exit 1
    fi
    "${CONTAINER_EXEC[@]}" -it "${CONTAINER_NAME}" bash
}

function cmd_exec() {
    if [ $# -eq 0 ]; then
        usage "No command specified for exec"
    fi
    if ! sudo podman ps --format '{{.Names}}' | grep -xFq "${CONTAINER_NAME}"; then
        echo "ERROR: Container '${CONTAINER_NAME}' is not running. Run '$(basename "$0") start' first."
        exit 1
    fi
    "${CONTAINER_EXEC[@]}" -it "${CONTAINER_NAME}" "$@"
}

function cmd_status() {
    if sudo podman ps --format '{{.Names}}' | grep -xFq "${CONTAINER_NAME}"; then
        echo "Container '${CONTAINER_NAME}': running"
        sudo podman exec "${CONTAINER_NAME}" \
            bash -c 'subscription-manager status 2>/dev/null && echo "Subscription: registered" || echo "Subscription: not registered"'
    elif sudo podman ps -a --format '{{.Names}}' | grep -xFq "${CONTAINER_NAME}"; then
        echo "Container '${CONTAINER_NAME}': stopped"
    else
        echo "Container '${CONTAINER_NAME}': not found"
    fi
}

#
# MAIN
#
if [ $# -lt 1 ]; then
    usage "No command specified"
fi

command="$1"
shift

case "${command}" in
    setup)  resolve_source; cmd_setup ;;
    start)  resolve_source; cmd_start ;;
    stop)   resolve_names; cmd_stop ;;
    delete) resolve_names; cmd_delete ;;
    shell)  resolve_names; cmd_shell ;;
    exec)   resolve_names; cmd_exec "$@" ;;
    status) resolve_names; cmd_status ;;
    *)      usage "Unknown command: ${command}" ;;
esac
