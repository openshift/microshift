#!/bin/bash
set -euo pipefail

OC_CMD="sudo -i oc --kubeconfig /var/lib/microshift/resources/kubeadmin/kubeconfig"

PRI_NODE_HOST=
PRI_NODE_ADDR=
SEC_NODE_HOST=
SEC_NODE_ADDR=

function usage() {
    echo "Usage: $(basename "$0") <primary_host_name> <primary_host_ip> <secondary_host_name> <secondary_host_ip>"
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

    # Configure the MicroShift host name
    sudo hostnamectl hostname "${SEC_NODE_HOST}"

    # Update /etc/hosts to resolve primary and secondary host names if not already resolved
    if [ "$(getent ahostsv4 "${PRI_NODE_HOST}" | head -1 | awk '{print $1}')" != "${PRI_NODE_ADDR}" ] ; then
        echo "${PRI_NODE_ADDR} ${PRI_NODE_HOST}" | sudo tee -a /etc/hosts &>/dev/null
    fi
    if [ "$(getent ahostsv4 "${SEC_NODE_HOST}" | head -1 | awk '{print $1}')" != "${SEC_NODE_ADDR}" ] ; then
        echo "${SEC_NODE_ADDR} ${SEC_NODE_HOST}" | sudo tee -a /etc/hosts &>/dev/null
    fi
}

function configure_microshift() {
    # Clean the current MicroShift configuration and stop the service
    echo 1 | sudo microshift-cleanup-data --all --keep-images

    # Run OVN initialization script
    sleep 5
    sudo systemctl start --wait microshift-ovs-init.service

    # OVN-K expects br-ex to have IP address assigned, add dummy IP to br-ex.
    if ! ip addr show br-ex 2>/dev/null | grep -q '10.44.0.0/32'; then
        sudo ip addr add 10.44.0.0/32 dev br-ex
    fi

    # Stop and unload the kubelet service
    if [ "$(systemctl is-active kubelet.service)" = "active" ] ; then
        sudo systemctl stop --now kubelet
    fi
    sudo systemctl reset-failed kubelet || true
    # Make sure the kubelet process is terminated
    sudo pkill -9 --exact kubelet || true
    until ! pidof kubelet &>/dev/null ; do
        sleep 1
    done
    # Clean up the old kubelet data
    for dir in $(mount | awk '{print $3}' | grep ^/var/lib/kubelet/) ; do
        sudo umount "${dir}"
    done
    sudo rm -rf /var/lib/kubelet
    # Remove the kubelet service unit
    sudo find /run/systemd -name kubelet.service -exec rm -f {} \;
    sudo systemctl daemon-reload
}

function configure_kubelet() {
    # Download the kubelet executable

    # Checksums can be obtained from https://www.downloadkubernetes.com/
    # or by downloading a "${url}.sha256" file (see below for ${url}). For example:
    # version=1.32.4; for kube_arch in amd64 arm64; do echo "${kube_arch}: $(curl -L https://dl.k8s.io/release/v${version}/bin/linux/${kube_arch}/kubelet.sha256 2>/dev/null)"; done
    local -r version="1.32.4"
    local kube_arch="amd64"
    local kube_hash="3e0c265fe80f3ea1b7271a00879d4dbd5e6ea1e91ecf067670c983e07c33a6f4"

    case $(uname -m) in
        x86_64)
            ;;
        aarch64)
            kube_arch="arm64"
            kube_hash="91117b71eb2bb3dd79ec3ed444e058a347349108bf661838f53ee30d2a0ff168"
            ;;
        *)
            echo "Unsupported kubelet architecture $(uname -m)"
            exit 1
    esac

    local -r url="https://dl.k8s.io/release/v${version}/bin/linux/${kube_arch}/kubelet"

    curl -sLO "${url}" --output-dir "${KUBELET_HOME}"
    cat <<EOF > "${KUBELET_HOME}/kubelet.sha256"
${kube_hash} ${KUBELET_HOME}/kubelet
EOF
    sha256sum --check "${KUBELET_HOME}/kubelet.sha256"
    chmod +x "${KUBELET_HOME}/kubelet"

    # OVN requires kubeconfig at this path
    # It must be a hard link or copy to be accessed from the container
    sudo mkdir -p /var/lib/microshift/resources/kubeadmin
    sudo ln "${KUBELET_HOME}/kubeconfig-${PRI_NODE_HOST}" /var/lib/microshift/resources/kubeadmin/kubeconfig
    # Remove the old kubelet configuration file so that it is recreated
    sudo rm -f "${KUBELET_HOME}/kubeconfig"

    # LVMS vg-manager requires presence of the lvmd.yaml file at a specific location
    sudo mkdir -p /var/lib/microshift/lvms
    sudo ln "${KUBELET_HOME}/lvmd-${PRI_NODE_HOST}.yaml" /var/lib/microshift/lvms/lvmd.yaml

    # Start crio & kubelet
    sudo systemctl enable --now crio
    sudo systemd-run --unit=kubelet --description="Kubelet" \
        --property=Environment="PATH=/sbin:/bin:/usr/sbin:/usr/bin:/opt/bin" \
        "${KUBELET_HOME}"/kubelet \
        --container-runtime-endpoint=/var/run/crio/crio.sock \
        --resolv-conf=/etc/resolv.conf \
        --rotate-certificates=true \
        --kubeconfig="${KUBELET_HOME}/kubeconfig" \
        --lock-file=/var/run/lock/kubelet.lock \
        --exit-on-lock-contention \
        --fail-swap-on=false \
        --max-pods=250 \
        --cgroup-driver=systemd \
        --tls-cert-file="${KUBELET_HOME}/kubelet-${SEC_NODE_HOST}.crt" \
        --tls-private-key-file="${KUBELET_HOME}/kubelet-${SEC_NODE_HOST}.key" \
        --bootstrap-kubeconfig="${KUBELET_HOME}/kubeconfig-${PRI_NODE_HOST}" \
        --cluster-dns=10.43.0.10 \
        --cluster-domain=cluster.local
}

function configure_node() {
    local service_ready=false

    echo "Waiting for MicroShift nodes to be ready"
    for _ in $(seq 300) ; do
        if ${OC_CMD} wait --for=condition=Ready nodes "${PRI_NODE_HOST}" --timeout=0s &>/dev/null ; then
            if ${OC_CMD} wait --for=condition=Ready nodes "${SEC_NODE_HOST}" --timeout=0s &>/dev/null ; then
                service_ready=true
                break
            fi
        fi
        echo -n "."
        sleep 1
    done
    echo

    if ! ${service_ready} ; then
        echo "Error: timed out waiting for MicroShift nodes to be ready"
        exit 1
    fi

    # Check all the nodes have the same kubelet version
    if ! ${OC_CMD} get node -o json | jq -e '[.items[].status.nodeInfo.kubeletVersion] | unique | length == 1' > /dev/null; then
        echo "Error: kubelet versions do not match"
        exit 1
    fi

    # Labeling the second node as a worker
    ${OC_CMD} label nodes "${SEC_NODE_HOST}" node-role.kubernetes.io/worker=
}

function wait_namespace_resources_ready() {
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
    sudo microshift healthcheck -v=2 --timeout="${wait_secs}s" --custom "${custom_json}"
}

#
# Main function
#
if [ $# -ne 4 ] ; then
    usage
fi
PRI_NODE_HOST=$1
PRI_NODE_ADDR=$2
SEC_NODE_HOST=$3
SEC_NODE_ADDR=$4

# Verify input file existence
KUBELET_HOME="${HOME}"
for file in "kubeconfig-${PRI_NODE_HOST}" "kubelet-${SEC_NODE_HOST}".{key,crt} "lvmd-${PRI_NODE_HOST}.yaml"; do
    if [ ! -e "${file}" ] ; then
        echo "The kubelet input file '${file}' is missing"
        exit 1
    fi
done

# Configure system for the multinode environment
configure_system

# Configure MicroShift for the multinode environment
configure_microshift

# Configure kubelet for the multinode environment
configure_kubelet

# Configure second node the multinode environment
configure_node

# Wait for all core namespace resources to be ready
wait_namespace_resources_ready

echo
echo "Done"
