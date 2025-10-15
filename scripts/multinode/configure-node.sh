#!/bin/bash
set -euo pipefail

# Variables for node configuration
NODE_ADDR=""
BOOTSTRAP_KUBECONFIG=""

function usage() {
    echo "This script configures a node to run a MicroShift cluster."
    echo "Optionally, it can also be used to configure MicroShift to join an existing cluster."
    echo "Usage: $(basename "$0") [OPTIONS]"
    echo "Options:"
    echo "  --bootstrap-kubeconfig PATH    Path to kubeconfig file for joining existing cluster (optional)"
    echo "  -h, --help                     Show this help message"
    exit 1
}

function configure_system() {
    # TODO: Edit firewall rules instead of stopping firewall
    sudo systemctl stop firewalld
    sudo systemctl disable firewalld

    sudo systemctl stop greenboot-healthcheck
    sudo systemctl reset-failed greenboot-healthcheck
    sudo systemctl disable greenboot-healthcheck
}

function configure_microshift() {
    # Clean the current MicroShift configuration and stop the service
    echo 1 | sudo microshift-cleanup-data --all --keep-images

    get_node_ip_from_config

    # Configure MicroShift to disable telemetry
    cat <<EOF | sudo tee /etc/microshift/config.d/multinode.yaml &>/dev/null
apiServer:
  subjectAltNames:
  - ${NODE_ADDR}
telemetry:
  status: Disabled
EOF

    sudo mkdir -p /etc/systemd/system/microshift.service.d
    cat <<EOF | sudo tee /etc/systemd/system/microshift.service.d/multinode.conf &>/dev/null
[Service]
# Clear previous ExecStart, otherwise systemd would try to run both.
ExecStart=
ExecStart=microshift run --multinode
EOF
    sudo systemctl daemon-reload
    sudo systemctl enable microshift.service
}

function start_microshift() {
    sudo systemctl start microshift.service
}

function run_add_node_commands() {
    if ! sudo microshift add-node --kubeconfig="${BOOTSTRAP_KUBECONFIG}"; then
        echo "Error: Failed to add node using kubeconfig: ${BOOTSTRAP_KUBECONFIG}"
        exit 1
    fi
    echo "Successfully added node using kubeconfig: ${BOOTSTRAP_KUBECONFIG}"
}

function get_node_ip_from_config() {
    # Extract nodeIP from MicroShift running configuration and store in global variable
    local node_ip=""

    # Use microshift show-config to get the IP address of the node
    node_ip=$(sudo microshift show-config 2>/dev/null | awk '/^\s*nodeIP\s*:/ {print $NF; exit}')

    if [ -z "${node_ip}" ]; then
        echo "Warning: nodeIP not found in MicroShift config"
        exit 1
    fi

    NODE_ADDR="${node_ip}"
}

function copy_bootstrap_kubeconfig() {
    local kubeconfig_source="/var/lib/microshift/resources/kubeadmin/${NODE_ADDR}/kubeconfig"
    local kubeconfig_dest="${HOME}/kubeconfig-bootstrap"

    if ! sudo test -f "${kubeconfig_source}"; then
        echo "Error: Kubeconfig file not found at ${kubeconfig_source}"
        exit 1
    fi

    if sudo cp "${kubeconfig_source}" "${kubeconfig_dest}"; then
        sudo chown "$(whoami):" "${kubeconfig_dest}"
        echo "Kubeconfig copied successfully to ${kubeconfig_dest}"
    else
        echo "Error: Failed to copy kubeconfig file"
        exit 1
    fi
}

function run_healthcheck() {
    if ! sudo systemctl start greenboot-healthcheck; then
        echo "Error: Failed to start greenboot-healthcheck service"
        exit 1
    fi

    greenboot_status=$(systemctl show -p Result --value greenboot-healthcheck)
    if [ "$greenboot_status" != "success" ]; then
        echo "Error: greenboot-healthcheck did not complete successfully (Result: $greenboot_status)"
        exit 1
    fi
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --bootstrap-kubeconfig)
            if [ $# -lt 2 ]; then
                echo "Error: --bootstrap-kubeconfig requires an argument"
                usage
            fi
            BOOTSTRAP_KUBECONFIG="$2"
            shift 2
            ;;
        -h|--help)
            usage
            ;;
        *)
            echo "Error: Unknown option '$1'"
            usage
            ;;
    esac
done

# Validate BOOTSTRAP_KUBECONFIG if provided
if [ -n "${BOOTSTRAP_KUBECONFIG}" ]; then
    if [ ! -f "${BOOTSTRAP_KUBECONFIG}" ]; then
        echo "Error: Bootstrap kubeconfig file '${BOOTSTRAP_KUBECONFIG}' does not exist"
        exit 1
    fi
    echo "Using bootstrap kubeconfig: ${BOOTSTRAP_KUBECONFIG}"
fi

configure_system
configure_microshift
if [ -n "${BOOTSTRAP_KUBECONFIG}" ]; then
    run_add_node_commands
    run_healthcheck
else
    start_microshift
fi
echo
echo "Node configuration completed"
if [ ! -n "${BOOTSTRAP_KUBECONFIG}" ]; then
    copy_bootstrap_kubeconfig
    echo
    echo "To add other nodes to this cluster, copy the following kubeconfig file to other nodes:"
    echo "  ${HOME}/kubeconfig-bootstrap"
    echo
    echo "Then run the following command on each node you want to add:"
    echo "  $(basename "$0") --bootstrap-kubeconfig /path/to/kubeconfig"
fi
echo "Done"
