#!/bin/bash
set -euo pipefail

OC_CMD="sudo -i oc --kubeconfig /var/lib/microshift/resources/kubeadmin/kubeconfig"
KUBECTL_CMD="sudo -i kubectl --kubeconfig /var/lib/microshift/resources/kubeadmin/kubeconfig"

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
    sudo hostnamectl hostname "${PRI_NODE_HOST}"

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

    # Configure MicroShift to advertise apiserver in the node IP
    cat <<EOF | sudo tee /etc/microshift/config.yaml &>/dev/null
apiServer:
  advertiseAddress: ${PRI_NODE_ADDR}
EOF
}

function wait_for_pod_ready() {
    local pod_namespace=$1
    local pod_name=$2
    local service_ready=false

    echo "Waiting for MicroShift ${pod_name}@${pod_namespace} pod to be ready"
    for _ in $(seq 300) ; do
        if ${OC_CMD} wait --timeout=0s --for=condition=ready pod -n "${pod_namespace}" -l app="${pod_name}" &>/dev/null ; then
            service_ready=true
            break
        fi
        echo -n "."
        sleep 1
    done
    echo

    if ! ${service_ready} ; then
        echo "Error: timed out waiting for MicroShift ${pod_name}@${pod_namespace} pod to be ready"
        exit 1
    fi
}

function generate_service_certs() {
    local -r cfssl=$(mktemp /tmp/cfssl.XXXXXXXXXX)
    local -r cfssl_json=$(mktemp /tmp/cfssl_json.XXXXXXXXXX)
    local -r cfssl_sha=$(mktemp /tmp/cfssl_sha.XXXXXXXXXX)
    local -r csr_file=$(mktemp /tmp/csrfile.XXXXXXXXXX)
    local -r kubelet_csr=$(mktemp /tmp/kubelet.XXXXXXXXXX)
    local cfssl_arch=""

    # Cleanup temporary files on exit
    # shellcheck disable=SC2064
    trap "rm -f ${cfssl} ${cfssl_json} ${cfssl_sha} ${csr_file} ${kubelet_csr}*" EXIT

    # Install cfssl utilities
    declare -A cfssl_map
    case "$(uname -m)" in
        x86_64)
            cfssl_arch="amd64"
            cfssl_map[cfssl]="b947d073e677189f8533704c44b2b1eae4042f5cefd2b8347d4d9b4c6a5008cf"
            cfssl_map[cfssl_json]="d7c52a815f96ebd4fc857b012cee70b44751edabb55ae60c4b743ee09e67f4de"
            ;;
        aarch64)
            cfssl_arch="arm64"
            cfssl_map[cfssl]="453495690f9b4e811d195d1f214ae58ad281e2e50e6dc3ffb19a8c58ddbc8a51"
            cfssl_map[cfssl_json]="4f68110f5a88a8b60382ff6b96008f55714636c31e5a06ab9c60855a1bd9bf47"
            ;;
        *)
            echo "Unsupported cfssl architecture $(uname -m)"
            exit 1
    esac

    curl -s -L -o "${cfssl}"      "https://github.com/cloudflare/cfssl/releases/download/v1.6.4/cfssl_1.6.4_linux_${cfssl_arch}"
    curl -s -L -o "${cfssl_json}" "https://github.com/cloudflare/cfssl/releases/download/v1.6.4/cfssljson_1.6.4_linux_${cfssl_arch}"
    cat <<EOF > "${cfssl_sha}"
${cfssl_map[cfssl]} ${cfssl}
${cfssl_map[cfssl_json]} ${cfssl_json}
EOF
    sha256sum --check "${cfssl_sha}"
    chmod +x "${cfssl}" "${cfssl_json}"

    # Generate serving certs for secondary node's kubelet
    cat <<EOF > "${csr_file}"
{
"CN": "system:node:${SEC_NODE_ADDR}",
"key": {
    "algo": "rsa",
    "size": 2048
},
"hosts": [
    "${SEC_NODE_HOST}",
    "${SEC_NODE_ADDR}"
],
"names": [
    {
        "O": "system:nodes"
    }
]
}
EOF

    # Generate and apply the certificate
    ${cfssl} genkey "${csr_file}" | ${cfssl_json} -bare "${kubelet_csr}"

    ${OC_CMD} apply -f - <<EOF
apiVersion: certificates.k8s.io/v1
kind: CertificateSigningRequest
metadata:
  name: csr-kubelet
spec:
  request: $(base64 < "${kubelet_csr}.csr" | tr -d '\n')
  signerName: kubernetes.io/kubelet-serving
  usages:
  - digital signature
  - key encipherment
  - server auth
EOF

    # Approve and extract the certificate
    ${KUBECTL_CMD} certificate approve csr-kubelet
    ${KUBECTL_CMD} get csr csr-kubelet -o jsonpath='{.status.certificate}' | base64 --decode > "${KUBELET_HOME}/kubelet-${SEC_NODE_HOST}.crt"
    cp "${kubelet_csr}-key.pem" "${KUBELET_HOME}/kubelet-${SEC_NODE_HOST}.key"

    # Copy the bootstrap kube configuration files
    sudo cp "/var/lib/microshift/resources/kubeadmin/${PRI_NODE_HOST}/kubeconfig" "${KUBELET_HOME}/kubeconfig-${PRI_NODE_HOST}"
    sudo chown "$(whoami)." "${KUBELET_HOME}/kubeconfig-${PRI_NODE_HOST}"

    # Copy lvmd configuration files for the second node
    sudo cp /var/lib/microshift/lvms/lvmd.yaml "${KUBELET_HOME}/lvmd-${PRI_NODE_HOST}.yaml"
    sudo chown "$(whoami)." "${KUBELET_HOME}/lvmd-${PRI_NODE_HOST}.yaml"
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

KUBELET_HOME="${HOME}"
if [ ! -w "${KUBELET_HOME}" ] ; then
    echo "The ${KUBELET_HOME} directory is not writable"
    exit 1
fi

# Configure system for the multinode environment
configure_system

# Configure MicroShift for the multinode environment
configure_microshift

# Run MicroShift in the multinode mode
sudo mkdir -p /etc/systemd/system/microshift.service.d
cat <<EOF | sudo tee /etc/systemd/system/microshift.service.d/multinode.conf &>/dev/null
[Service]
# Clear previous ExecStart, otherwise systemd would try to run both.
ExecStart=
ExecStart=microshift run --multinode
EOF
sudo systemctl daemon-reload
sudo systemctl start microshift.service

# Wait for the service-ca pod to be ready
wait_for_pod_ready "openshift-service-ca" "service-ca"

# Generate the certificates for the multinode environment
generate_service_certs

# Print the file names to be copied to the secondary host
echo
echo "Copy the following files to the ${SEC_NODE_HOST} host"
ls -1 "${KUBELET_HOME}/kubeconfig-${PRI_NODE_HOST}"
ls -1 "${KUBELET_HOME}/kubelet-${SEC_NODE_HOST}".{key,crt}
ls -1 "${KUBELET_HOME}/lvmd-${PRI_NODE_HOST}.yaml"

echo
echo "Done"
