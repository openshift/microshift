#!/bin/bash

set -xeuo pipefail

# 4 cores for management pods
MANAGEMENT_CPUSET="0-3"

# Rest for the workloads
WORKLOAD_CPUSET="4-23"

# c5.metal's NUMA configuration is:
#   NUMA node0 CPU(s):     0-23,48-71
#   NUMA node1 CPU(s):     24-47,72-95
# Let's offline following CPUs to only run test on a single NUMA node with SMT disabled:
# - 48-71 - SMT part of the NUMA node0
# - 24-47,72-95 which is NUMA node1
OFFLINE_CPUSET="24-95"

# Kubelet configuration - most important stuff:
# - CPU Manager "static" policy
# - Memory Manager "Static" policy
# - Topology Manager "single-numa-node" policy
# - Reserved System CPUs equal to our management pods CPUs
sudo mkdir -p /etc/microshift
sudo tee /etc/microshift/config.yaml << EOF
kubelet:
  cpuManagerPolicy: static
  cpuManagerPolicyOptions:
    full-pcpus-only: "true"
  cpuManagerReconcilePeriod: 5s
  memoryManagerPolicy: Static
  topologyManagerPolicy: single-numa-node
  reservedSystemCPUs: ${MANAGEMENT_CPUSET}
  reservedMemory:
  - limits:
      memory: 1100Mi
    numaNode: 0
  kubeReserved:
    memory: 500Mi
  systemReserved:
    memory: 500Mi
  evictionHard:
    imagefs.available: 15%
    memory.available: 100Mi
    nodefs.available: 10%
    nodefs.inodesFree: 5%
  evictionPressureTransitionPeriod: 0s
EOF

# Configure microshift-baseline TuneD profile:
#  - isolate the cores for the workload
#  - set up 1000 of 2M hugepages (no particular purpose)
#  - disable 2nd NUMA node and SMT cores for 1st NUMA node
sudo mkdir -p /etc/tuned/
sudo tee /etc/tuned/microshift-baseline-variables.conf << EOF
isolated_cores=${WORKLOAD_CPUSET}
hugepages_size=2M
hugepages=1000
additional_args=
offline_cpu_set=${OFFLINE_CPUSET}
EOF

# Configure workload partitioning: pin MicroShift Pods to $MANAGEMENT_CPUSET
sudo mkdir -p /etc/crio/crio.conf.d/
sudo tee /etc/crio/crio.conf.d/20-microshift-wp.conf << EOF
[crio.runtime]
infra_ctr_cpuset = "${MANAGEMENT_CPUSET}"

[crio.runtime.workloads.management]
activation_annotation = "target.workload.openshift.io/management"
annotation_prefix = "resources.workload.openshift.io"
resources = { "cpushares" = 0, "cpuset" = "${MANAGEMENT_CPUSET}" }
EOF

sudo mkdir -p /etc/kubernetes/
sudo tee /etc/kubernetes/openshift-workload-pinning << EOF
{
  "management": {
    "cpuset": "${MANAGEMENT_CPUSET}"
  }
}
EOF

dropins=(
    "/etc/systemd/system/ovs-vswitchd.service.d/microshift-cpuaffinity.conf"
    "/etc/systemd/system/ovsdb-server.service.d/microshift-cpuaffinity.conf"
    "/etc/systemd/system/crio.service.d/microshift-cpuaffinity.conf"
    "/etc/systemd/system/microshift.service.d/microshift-cpuaffinity.conf"
)
for path in "${dropins[@]}"; do
    sudo mkdir -p "$(dirname "${path}")"
    sudo tee "${path}" << EOF
[Service]
CPUAffinity=${MANAGEMENT_CPUSET}
EOF
done
