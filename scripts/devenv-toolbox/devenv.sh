#!/bin/bash
set -eou pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOTDIR="$( cd "${SCRIPTDIR}/../../" && pwd )"
WORKTREE_BASE="${ROOTDIR}/.worktrees"

CONTAINERFILE="${SCRIPTDIR}/Containerfile.toolbox"

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
    TOOLBOX_NAME="microshift-toolbox-${BRANCH_TAG}"
    IMAGE_NAME="microshift-toolbox:${BRANCH_TAG}"
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
    echo "  setup    Build the toolbox image"
    echo "  start    Create and configure the toolbox"
    echo "  stop     Stop the toolbox (no-op, toolbox has no stop)"
    echo "  delete   Delete the toolbox"
    echo "  shell    Open a shell in the toolbox"
    echo "  exec     Execute a command in the toolbox"
    echo "  status   Show the toolbox status"
    echo ""
    echo "Environment variables:"
    echo "  PULL_SECRET          Path to pull secret (default: ~/.pull-secret.json)"
    echo "  RHSM_ORG             Red Hat subscription org ID (required for 'setup' and 'start')"
    echo "  RHSM_ACTIVATION_KEY  Red Hat subscription activation key (required for 'setup' and 'start')"
    echo "  DEVENV_BRANCH        Build for a different branch (creates a worktree)"
    echo ""
    echo "Examples:"
    echo "  $(basename "$0") setup                                      # build using current branch"
    echo "  $(basename "$0") start                                      # create and configure the toolbox"
    echo "  DEVENV_BRANCH=release-4.21 $(basename "$0") setup           # build for release-4.21"
    echo "  DEVENV_BRANCH=release-4.21 $(basename "$0") shell           # shell into release-4.21 toolbox"

    [ -n "$1" ] && echo -e "\nERROR: $1"
    exit 1
}

function check_prerequisites() {
    if ! command -v toolbox &>/dev/null; then
        echo "ERROR: toolbox is not installed"
        exit 1
    fi
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

function toolbox_exists() {
    podman container exists "${TOOLBOX_NAME}" &>/dev/null
}

function toolbox_running() {
    podman ps --format '{{.Names}}' | grep -xFq "${TOOLBOX_NAME}"
}

function toolbox_exec() {
    toolbox run -c "${TOOLBOX_NAME}" "$@"
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

    echo "Building toolbox image '${IMAGE_NAME}' (RHEL ${RHEL_VERSION}) for branch '${DEVENV_BRANCH}'..."
    podman build "${build_args[@]}" "${ROOTDIR}"
}

function cmd_start() {
    check_prerequisites
    if [ -z "${RHSM_ORG:-}" ] || [ -z "${RHSM_ACTIVATION_KEY:-}" ]; then
        echo "ERROR: RHSM_ORG and RHSM_ACTIVATION_KEY env vars are required"
        exit 1
    fi

    # Already exists
    if toolbox_exists; then
        echo "Toolbox '${TOOLBOX_NAME}' already exists"
        return 0
    fi

    # Create toolbox from the custom image
    echo "Creating toolbox '${TOOLBOX_NAME}'..."
    toolbox create --image "${IMAGE_NAME}" "${TOOLBOX_NAME}"

    # Copy pull secret to expected locations
    toolbox_exec sudo mkdir -p /etc/crio
    toolbox_exec sudo cp "${PULL_SECRET}" /etc/crio/openshift-pull-secret

    # Configure git safe directory
    toolbox_exec sudo git config --system --add safe.directory "${SOURCE_DIR}"

    # Register subscription
    echo "Registering subscription..."
    toolbox_exec sudo subscription-manager register \
        --org="${RHSM_ORG}" \
        --activationkey="${RHSM_ACTIVATION_KEY}"

    # Run the release branch's configure-vm.sh
    echo "Running configure-vm.sh..."
    toolbox_exec bash -x "${SOURCE_DIR}/scripts/devenv-builder/configure-vm.sh" \
        --no-build --skip-dnf-update /etc/crio/openshift-pull-secret

    # Configure composer
    echo "Running manage_composer_config.sh..."
    toolbox_exec bash -x "${SOURCE_DIR}/test/bin/manage_composer_config.sh" create

    # Configure hypervisor for VM management
    echo "Running manage-vm.sh config and manage_hypervisor_config.sh create..."
    toolbox_exec bash -x -c "${SOURCE_DIR}/scripts/devenv-builder/manage-vm.sh config && ${SOURCE_DIR}/test/bin/manage_hypervisor_config.sh create"

    echo "Toolbox '${TOOLBOX_NAME}' ready"
    echo "Use '$(basename "$0") shell' to open a shell"
}

function cmd_stop() {
    if ! toolbox_exists; then
        echo "Toolbox '${TOOLBOX_NAME}' does not exist"
        return 0
    fi
    if ! toolbox_running; then
        echo "Toolbox '${TOOLBOX_NAME}' is not running"
        return 0
    fi
    podman stop "${TOOLBOX_NAME}"
    echo "Toolbox '${TOOLBOX_NAME}' stopped (running processes killed)"
}

function cmd_delete() {
    if ! toolbox_exists; then
        echo "Toolbox '${TOOLBOX_NAME}' does not exist"
        return 0
    fi
    toolbox rm -f "${TOOLBOX_NAME}"
    echo "Toolbox '${TOOLBOX_NAME}' deleted"
}

function cmd_shell() {
    if ! toolbox_exists; then
        echo "ERROR: Toolbox '${TOOLBOX_NAME}' does not exist. Run '$(basename "$0") start' first."
        exit 1
    fi
    toolbox enter "${TOOLBOX_NAME}"
}

function cmd_exec() {
    if [ $# -eq 0 ]; then
        usage "No command specified for exec"
    fi
    if ! toolbox_exists; then
        echo "ERROR: Toolbox '${TOOLBOX_NAME}' does not exist. Run '$(basename "$0") start' first."
        exit 1
    fi
    toolbox_exec "$@"
}

function cmd_status() {
    if toolbox_running; then
        echo "Toolbox '${TOOLBOX_NAME}': running"
    elif toolbox_exists; then
        echo "Toolbox '${TOOLBOX_NAME}': exists (stopped)"
    else
        echo "Toolbox '${TOOLBOX_NAME}': not found"
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
