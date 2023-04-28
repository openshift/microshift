#!/bin/bash
set -e

WORKER_NODE_HOSTNAME=${WORKER_NODE_HOSTNAME:-"worker-node"}
: "${WORKER_NODE_IP:=$(getent hosts "${WORKER_NODE_HOSTNAME}" | awk '{print $1}')}"
KUBELET_DIR=${KUBELET_DIR:-"/home/microshift/kubelet"}

remote() {
  if [ "$#" -lt 2 ]; then
    echo "Usage: remote remote_host command [arg1 arg2 ...]"
    return 1
  fi

  remote_host="$1"
  shift

  ssh -q -t -t "microshift@${remote_host}" "$(printf "%q " "$@")"
}

# pre-approve worker node host key
ssh-keyscan "${WORKER_NODE_HOSTNAME}" >> ~/.ssh/known_hosts

### ON MASTER NODE

sudo systemctl stop microshift

# configure microshift to run in multinode mode
echo 'mtu: 1422' | sudo tee /etc/microshift/ovn.yaml
cat << EOF | sudo tee /etc/microshift/config.yaml
apiServer:
  subjectAltNames:
    - "10.43.0.1"
EOF

# TODO - microshift-etcd isn't happy if microshift is also a transient service unit
# sudo systemd-run --unit=microshift-multinode \
#   --description="Multinode MicroShift" \
#   --property="WorkingDirectory=/usr/bin" \
#   --property="User=root" \
#   --property="Type=notify" \
#   --property="Delegate=yes" \
#   --property="CPUAccounting=yes" \
#   --property="BlockIOAccounting=yes" \
#   --property="MemoryAccounting=yes" \
#   --property="LimitNOFILE=1048576" \
#   --property="TimeoutStartSec=2m" \
#   /usr/bin/microshift run --multinode

sudo cp -n /usr/lib/systemd/system/microshift.service /usr/lib/systemd/system/microshift.service.backup
cat << EOF | sudo tee /usr/lib/systemd/system/microshift.service
[Unit]
Description=MicroShift
Wants=network-online.target crio.service openvswitch.service microshift-ovs-init.service
After=network-online.target crio.service openvswitch.service microshift-ovs-init.service

# Control shutdown order by declaring this service to start Before the kubepods.slice
# transient systemd unit; this makes system shutdown delay MicroShift shutdown until
# all the pod containers are down. This is important because some services need to talk
# to the MicroShift API during shutdown (i.e. releasing leader election locks or cleaning
# up other resources) MicroShift restart or manual stop will not stop the kubepods.
Before=kubepods.slice

[Service]
WorkingDirectory=/usr/bin/
ExecStart=microshift run --multinode
Restart=always
User=root
Type=notify
Delegate=yes
CPUAccounting=yes
BlockIOAccounting=yes
MemoryAccounting=yes
LimitNOFILE=1048576
TimeoutStartSec=2m

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl start microshift

# generate serving certs for worker node's kubelet

go install github.com/cloudflare/cfssl/cmd/...@latest
export PATH="${PATH}:${HOME}/go/bin"

cat << EOF >csr.json
{
  "CN": "system:node:${WORKER_NODE_IP}",
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "hosts": [
    "${WORKER_NODE_HOSTNAME}",
    "${WORKER_NODE_IP}"
  ],
  "names": [
    {
      "O": "system:nodes"
    }
  ]
}
EOF

cfssl genkey csr.json | cfssljson -bare kubelet-"${WORKER_NODE_HOSTNAME}"

cat <<EOF > csr.yaml
apiVersion: certificates.k8s.io/v1
kind: CertificateSigningRequest
metadata:
  name: csr-kubelet
spec:
  request: $(base64 < kubelet-"${WORKER_NODE_HOSTNAME}".csr | tr -d '\n')
  signerName: kubernetes.io/kubelet-serving
  usages:
  - digital signature
  - key encipherment
  - server auth
EOF

oc apply -f csr.yaml

kubectl certificate approve csr-kubelet
kubectl get csr csr-kubelet -o jsonpath='{.status.certificate}' | base64 --decode > kubelet.crt

remote "${WORKER_NODE_HOSTNAME}" mkdir -p "${KUBELET_DIR}"
scp kubelet-"${WORKER_NODE_HOSTNAME}"-key.pem microshift@"${WORKER_NODE_HOSTNAME}":"${KUBELET_DIR}"/kubelet.key
scp kubelet.crt microshift@"${WORKER_NODE_HOSTNAME}":"${KUBELET_DIR}"/kubelet.crt
sudo cp /var/lib/microshift/resources/kubeadmin/"${HOSTNAME}"/kubeconfig ~/bootstrap-kubeconfig
sudo chmod a+r ~/bootstrap-kubeconfig
scp ~/bootstrap-kubeconfig microshift@"${WORKER_NODE_HOSTNAME}":"${KUBELET_DIR}"/bootstrap-kubeconfig

# stop firewall on master (to allow connections to port 9642 from worker-node and allow OVN to work properly)
# TODO edit firewall rules instead of stopping firewall
sudo systemctl stop firewalld;
sudo systemctl disable firewalld;

### ON WORKER NODE

# stop firewall on worker (to allow connections to kubelet from master)
# TODO edit firewall rules instead of stopping firewall
remote "${WORKER_NODE_HOSTNAME}" sudo systemctl stop firewalld;
remote "${WORKER_NODE_HOSTNAME}" sudo systemctl disable firewalld;

# download kubelet
remote "${WORKER_NODE_HOSTNAME}" curl -LO https://dl.k8s.io/release/v1.26.1/bin/linux/amd64/kubelet --output-dir "${KUBELET_DIR}"
# shellcheck disable=SC2029
echo "8b99dd73f309ca1ac4005db638e82f949ffcfb877a060089ec0e729503db8198  ${KUBELET_DIR}/kubelet" | ssh microshift@"${WORKER_NODE_HOSTNAME}" "cat > ${KUBELET_DIR}/kubelet.sha256"
remote "${WORKER_NODE_HOSTNAME}" sha256sum --check "${KUBELET_DIR}"/kubelet.sha256
remote "${WORKER_NODE_HOSTNAME}" chmod +x "${KUBELET_DIR}"/kubelet

# wanted at this path by OVN
remote "${WORKER_NODE_HOSTNAME}" sudo mkdir -p /var/lib/microshift/resources/kubeadmin
remote "${WORKER_NODE_HOSTNAME}" sudo ln "${KUBELET_DIR}"/bootstrap-kubeconfig /var/lib/microshift/resources/kubeadmin/kubeconfig

# start crio & kubelet
remote "${WORKER_NODE_HOSTNAME}" sudo systemctl enable --now crio

# disable selinux TODO remove once selinux is working properly again
remote "${WORKER_NODE_HOSTNAME}" sudo setenforce 0

remote "${WORKER_NODE_HOSTNAME}" sudo systemd-run --unit=kubelet --description="Kubelet" \
  --property=Environment="PATH=/sbin:/bin:/usr/sbin:/usr/bin:/opt/bin" \
  "${KUBELET_DIR}"/kubelet \
    --container-runtime-endpoint=/var/run/crio/crio.sock \
    --resolv-conf=/etc/resolv.conf \
    --rotate-certificates=true \
    --kubeconfig="${KUBELET_DIR}"/kubeconfig \
    --lock-file=/var/run/lock/kubelet.lock \
    --exit-on-lock-contention \
    --fail-swap-on=false \
    --cgroup-driver=systemd \
    --tls-cert-file="${KUBELET_DIR}"/kubelet.crt \
    --tls-private-key-file="${KUBELET_DIR}"/kubelet.key \
    --bootstrap-kubeconfig="${KUBELET_DIR}"/bootstrap-kubeconfig \
    --cluster-dns=10.43.0.10 \
    --cluster-domain=cluster.local
