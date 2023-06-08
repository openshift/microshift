#!/bin/bash
set -euo pipefail

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
    echo 1 | sudo microshift-cleanup-data --all

    # Run OVN initialization script
    sleep 5
    sudo systemctl start microshift-ovs-init.service

    # Stop and unload the kubelet service, cleaning up its old data
    if [ "$(systemctl is-active kubelet.service)" = "active" ] ; then
        sudo systemctl stop --now kubelet
    fi
    sudo systemctl reset-failed kubelet || true
    sudo pkill -9 --exact kubelet       || true
    sudo rm -rf /var/lib/kubelet &> /dev/null || true

    sudo find /run/systemd -name kubelet.service -exec rm -f {} \;
    sudo systemctl daemon-reload
}

function configure_kubelet() {
    # Download the kubelet executable
    local kube_arch="amd64"
    local kube_hash="cb2845fff0ce41c400489393da73925d28fbee54cfeb7834cd4d11e622cbd3a7"

    case $(uname -m) in
        x86_64)
            ;;
        aarch64)
            kube_arch="arm64"
            kube_hash="dbb09d297d924575654db38ed2fc627e35913c2d4000c34613ac6de4995457d0"
            ;;
        *)
            echo "Unsupported kubelet architecture $(uname -m)"
            exit 1
    esac

    curl -sLO "https://dl.k8s.io/release/v1.27.1/bin/linux/${kube_arch}/kubelet" --output-dir "${KUBELET_HOME}"
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
for file in "kubeconfig-${PRI_NODE_HOST}" "kubelet-${SEC_NODE_HOST}".{key,crt} ; do
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

echo
echo "Done"
