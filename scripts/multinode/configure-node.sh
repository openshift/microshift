#!/bin/bash
set -euo pipefail

OC_CMD="sudo -i oc --kubeconfig /var/lib/microshift/resources/kubeadmin/kubeconfig"

# Variables for node configuration
NODE_ADDR=""
BOOTSTRAP_KUBECONFIG="${BOOTSTRAP_KUBECONFIG:-}"

function usage() {
    echo "Usage: $(basename "$0")"
    echo "Environment variables:"
    echo "  BOOTSTRAP_KUBECONFIG    - Path to kubeconfig file for joining existing cluster (optional)"
    exit 1
}

function configure_system() {
    # Disable selinux
    # TODO: remove once selinux is working properly again
    sudo setenforce 0 || true

    # TODO: Edit firewall rules instead of stopping firewall
    sudo systemctl stop firewalld
    sudo systemctl disable firewalld

    # Greenboot checks are tuned for a single node
    sudo systemctl stop greenboot-healthcheck
    sudo systemctl reset-failed greenboot-healthcheck
    sudo systemctl disable greenboot-healthcheck
}

function configure_microshift() {
    # Clean the current MicroShift configuration and stop the service
    echo 1 | sudo microshift-cleanup-data --all --keep-images

    get_node_ip_from_config

    # Configure MicroShift to disable telemetry
    cat <<EOF | sudo tee /etc/microshift/config.yaml &>/dev/null
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

    # Use microshift show-config to get the actual running configuration
    if command -v microshift >/dev/null 2>&1; then
        # Get the nodeIP from the show-config output
        node_ip=$(sudo microshift show-config 2>/dev/null | grep -E "^\s*nodeIP\s*:" | head -1 | sed 's/.*nodeIP\s*:\s*//' | tr -d ' \t\r\n')
    fi

    if [ -z "${node_ip}" ]; then
        echo "Warning: nodeIP not found in MicroShift config"
        exit 1
    fi

    NODE_ADDR="${node_ip}"
}

function copy_kubeconfig() {
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
    local -r wait_secs=600
    local -r custom_json=$(cat <<EOF
{
    "openshift-ovn-kubernetes": {
        "daemonsets": ["ovnkube-master", "ovnkube-node"]
    },
    "openshift-service-ca": {
        "deployments": ["service-ca"]
    },
    "openshift-ingress": {
        "deployments": ["router-default"]
    },
    "openshift-dns": {
        "daemonsets": ["dns-default", "node-resolver"]
    },
    "openshift-storage": {
        "deployments": ["lvms-operator"],
        "daemonsets": ["vg-manager"]
    },
    "kube-system": {
        "deployments": ["csi-snapshot-controller"]
    }
}
EOF
)
    if ! sudo microshift healthcheck -v=2 --timeout="${wait_secs}s" --custom "${custom_json}"; then
        echo "Error: MicroShift healthcheck failed"
        exit 1
    fi
}

#
# Main function
#
if [ -n "${BOOTSTRAP_KUBECONFIG}" ]; then
    if [ ! -f "${BOOTSTRAP_KUBECONFIG}" ]; then
        echo "Error: BOOTSTRAP_KUBECONFIG is set to '${BOOTSTRAP_KUBECONFIG}' but the file does not exist"
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
    copy_kubeconfig
    echo
    echo "To add other nodes to this cluster, copy the following kubeconfig file to other nodes:"
    echo "  ${HOME}/kubeconfig-bootstrap"
    echo
    echo "Then run the following command on each node you want to add:"
    echo "  BOOTSTRAP_KUBECONFIG=/path/to/kubeconfig $(basename "$0")"
fi
echo "Done"
