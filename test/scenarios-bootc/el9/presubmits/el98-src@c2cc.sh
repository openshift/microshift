#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# Cluster A (host1): default MicroShift CIDRs
CLUSTER_A_POD_CIDR="10.42.0.0/16"
CLUSTER_A_SVC_CIDR="10.43.0.0/16"
CLUSTER_A_DOMAIN="cluster-a.remote"

# Cluster B (host2): non-overlapping CIDRs
CLUSTER_B_POD_CIDR="10.45.0.0/16"
CLUSTER_B_SVC_CIDR="10.46.0.0/16"
CLUSTER_B_DOMAIN="cluster-b.remote"

configure_c2cc_hosts() {
    local -r host1_ip=$(get_vm_property host1 ip)
    local -r host2_ip=$(get_vm_property host2 ip)

    # host1 (Cluster A): C2CC config pointing to host2
    run_command_on_vm host1 "sudo mkdir -p /etc/microshift/config.d"
    run_command_on_vm host1 "sudo tee /etc/microshift/config.d/50-c2cc.yaml > /dev/null << EOF
clusterToCluster:
  remoteClusters:
  - nextHop: ${host2_ip}
    clusterNetwork:
    - ${CLUSTER_B_POD_CIDR}
    serviceNetwork:
    - ${CLUSTER_B_SVC_CIDR}
    domain: ${CLUSTER_B_DOMAIN}
EOF"

    # host2 (Cluster B): custom CIDRs + C2CC config pointing to host1
    run_command_on_vm host2 "sudo mkdir -p /etc/microshift/config.d"
    run_command_on_vm host2 "sudo tee /etc/microshift/config.d/50-c2cc.yaml > /dev/null << EOF
network:
  clusterNetwork:
  - ${CLUSTER_B_POD_CIDR}
  serviceNetwork:
  - ${CLUSTER_B_SVC_CIDR}
clusterToCluster:
  remoteClusters:
  - nextHop: ${host1_ip}
    clusterNetwork:
    - ${CLUSTER_A_POD_CIDR}
    serviceNetwork:
    - ${CLUSTER_A_SVC_CIDR}
    domain: ${CLUSTER_A_DOMAIN}
EOF"

    # Standard firewall setup on both hosts
    configure_vm_firewall host1
    configure_vm_firewall host2

    # Trust remote pod and service CIDRs
    run_command_on_vm host1 "sudo firewall-cmd --permanent --zone=trusted --add-source=${CLUSTER_B_POD_CIDR}"
    run_command_on_vm host1 "sudo firewall-cmd --permanent --zone=trusted --add-source=${CLUSTER_B_SVC_CIDR}"
    run_command_on_vm host1 "sudo firewall-cmd --reload"

    run_command_on_vm host2 "sudo firewall-cmd --permanent --zone=trusted --add-source=${CLUSTER_A_POD_CIDR}"
    run_command_on_vm host2 "sudo firewall-cmd --permanent --zone=trusted --add-source=${CLUSTER_A_SVC_CIDR}"
    run_command_on_vm host2 "sudo firewall-cmd --reload"

    # host2 CIDRs changed from defaults — clean up stale OVN/etcd state
    run_command_on_vm host2 "echo 1 | sudo microshift-cleanup-data --all --keep-images"

    # Restart MicroShift on both hosts (enable on host2 since cleanup-data disables it)
    run_command_on_vm host1 "sudo systemctl restart microshift"
    run_command_on_vm host2 "sudo systemctl enable --now microshift"

    # Wait for greenboot on both hosts
    local -r full_host1=$(full_vm_name host1)
    local -r full_host2=$(full_vm_name host2)
    if ! wait_for_greenboot "${full_host1}" "${host1_ip}"; then
        record_junit host1 "c2cc_greenboot" "FAILED"
        return 1
    fi
    record_junit host1 "c2cc_greenboot" "OK"

    if ! wait_for_greenboot "${full_host2}" "${host2_ip}"; then
        record_junit host2 "c2cc_greenboot" "FAILED"
        return 1
    fi
    record_junit host2 "c2cc_greenboot" "OK"
}

scenario_create_vms() {
    # TODO: Configure network.clusterNetwork and network.serviceNetwork on install, so after boot it'll be only c2cc config & restart (no cleanup)
    prepare_kickstart host1 kickstart-bootc.ks.template rhel98-bootc-source
    prepare_kickstart host2 kickstart-bootc.ks.template rhel98-bootc-source
    launch_vm rhel98-bootc --vmname host1
    launch_vm rhel98-bootc --vmname host2
}

scenario_remove_vms() {
    remove_vm host1
    remove_vm host2
}

scenario_run_tests() {
    if ! configure_c2cc_hosts; then
        return 1
    fi

    # Retrieve host2's kubeconfig
    local -r host2_ip=$(get_vm_property host2 ip)
    local -r host2_ssh_port=$(get_vm_property host2 ssh_port)
    local -r kubeconfig_b="${SCENARIO_INFO_DIR}/${SCENARIO}/kubeconfig-b"

    # Wait for host2 to be fully ready (run_tests only waits for host1)
    wait_for_microshift_to_be_ready host2

    run_command_on_vm host2 "sudo cp /var/lib/microshift/resources/kubeadmin/${host2_ip}/kubeconfig /tmp/kubeconfig-b && sudo chmod 644 /tmp/kubeconfig-b"
    scp -P "${host2_ssh_port}" "redhat@${host2_ip}:/tmp/kubeconfig-b" "${kubeconfig_b}"

    run_tests host1 \
        --variable "CLUSTER_A_POD_CIDR:${CLUSTER_A_POD_CIDR}" \
        --variable "CLUSTER_A_SVC_CIDR:${CLUSTER_A_SVC_CIDR}" \
        --variable "CLUSTER_A_DOMAIN:${CLUSTER_A_DOMAIN}" \
        --variable "CLUSTER_B_POD_CIDR:${CLUSTER_B_POD_CIDR}" \
        --variable "CLUSTER_B_SVC_CIDR:${CLUSTER_B_SVC_CIDR}" \
        --variable "CLUSTER_B_DOMAIN:${CLUSTER_B_DOMAIN}" \
        --variable "KUBECONFIG_B:${kubeconfig_b}" \
        suites/c2cc/
}
