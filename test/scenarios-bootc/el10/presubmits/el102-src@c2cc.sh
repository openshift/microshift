#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.
export TEST_RANDOMIZATION=none

# Cluster A (host1): default MicroShift CIDRs
CLUSTER_A_POD_CIDR="10.42.0.0/16"
CLUSTER_A_SVC_CIDR="10.43.0.0/16"
CLUSTER_A_DOMAIN="cluster-a.remote"

# Cluster B (host2): non-overlapping CIDRs
CLUSTER_B_POD_CIDR="10.45.0.0/16"
CLUSTER_B_SVC_CIDR="10.46.0.0/16"
CLUSTER_B_DOMAIN="cluster-b.remote"

# Cluster C (host3): non-overlapping CIDRs
CLUSTER_C_POD_CIDR="10.48.0.0/16"
CLUSTER_C_SVC_CIDR="10.49.0.0/16"
CLUSTER_C_DOMAIN="cluster-c.remote"

wait_for_greenboot_on_hosts() {
    local junit_label=$1
    local host
    for host in host1 host2 host3; do
        local host_ip full_host
        host_ip=$(get_vm_property "${host}" ip)
        full_host=$(full_vm_name "${host}")
        if ! wait_for_greenboot "${full_host}" "${host_ip}"; then
            record_junit "${host}" "${junit_label}" "FAILED"
            return 1
        fi
        record_junit "${host}" "${junit_label}" "OK"
    done
}

configure_c2cc_host() {
    local host=$1 remote_ip=$2 remote_pod_cidr=$3 remote_svc_cidr=$4 remote_domain=$5

    run_command_on_vm "${host}" "sudo mkdir -p /etc/microshift/config.d"
    run_command_on_vm "${host}" "sudo tee /etc/microshift/config.d/50-c2cc-${remote_domain}.yaml > /dev/null << EOF
clusterToCluster:
  remoteClusters:
  - nextHop: ${remote_ip}
    clusterNetwork:
    - ${remote_pod_cidr}
    serviceNetwork:
    - ${remote_svc_cidr}
    domain: ${remote_domain}
EOF"

    configure_vm_firewall "${host}"
    run_command_on_vm "${host}" "sudo firewall-cmd --permanent --zone=trusted --add-source=${remote_pod_cidr}"
    run_command_on_vm "${host}" "sudo firewall-cmd --permanent --zone=trusted --add-source=${remote_svc_cidr}"
    run_command_on_vm "${host}" "sudo firewall-cmd --reload"

    run_command_on_vm "${host}" "sudo systemctl restart microshift"
}

configure_c2cc_hosts() {
    local -r host1_ip=$(get_vm_property host1 ip)
    local -r host2_ip=$(get_vm_property host2 ip)
    local -r host3_ip=$(get_vm_property host3 ip)

    wait_for_greenboot_on_hosts "c2cc_pre_greenboot"

    configure_c2cc_host host1 "${host2_ip}" "${CLUSTER_B_POD_CIDR}" "${CLUSTER_B_SVC_CIDR}" "${CLUSTER_B_DOMAIN}"
    configure_c2cc_host host1 "${host3_ip}" "${CLUSTER_C_POD_CIDR}" "${CLUSTER_C_SVC_CIDR}" "${CLUSTER_C_DOMAIN}"
    configure_c2cc_host host2 "${host1_ip}" "${CLUSTER_A_POD_CIDR}" "${CLUSTER_A_SVC_CIDR}" "${CLUSTER_A_DOMAIN}"
    configure_c2cc_host host2 "${host3_ip}" "${CLUSTER_C_POD_CIDR}" "${CLUSTER_C_SVC_CIDR}" "${CLUSTER_C_DOMAIN}"
    configure_c2cc_host host3 "${host1_ip}" "${CLUSTER_A_POD_CIDR}" "${CLUSTER_A_SVC_CIDR}" "${CLUSTER_A_DOMAIN}"
    configure_c2cc_host host3 "${host2_ip}" "${CLUSTER_B_POD_CIDR}" "${CLUSTER_B_SVC_CIDR}" "${CLUSTER_B_DOMAIN}"

    wait_for_greenboot_on_hosts "c2cc_greenboot"
}

scenario_create_vms() {
    prepare_kickstart host1 kickstart-bootc.ks.template rhel102-bootc-source
    prepare_kickstart host2 kickstart-bootc.ks.template rhel102-bootc-source
    prepare_kickstart host3 kickstart-bootc.ks.template rhel102-bootc-source

    # Inject host2's and host3's non-default CIDRs into its kickstart config so MicroShift
    # boots with the correct network from the start (no cleanup-data needed).
    local -r host2_ks_dir="${SCENARIO_INFO_DIR}/${SCENARIO}/vms/host2"
    cat >> "${host2_ks_dir}/post-microshift.cfg" <<EOF
cat - >>/etc/microshift/config.yaml <<IEOF
network:
  clusterNetwork:
  - ${CLUSTER_B_POD_CIDR}
  serviceNetwork:
  - ${CLUSTER_B_SVC_CIDR}
IEOF
EOF
    local -r host3_ks_dir="${SCENARIO_INFO_DIR}/${SCENARIO}/vms/host3"
    cat >> "${host3_ks_dir}/post-microshift.cfg" <<EOF
cat - >>/etc/microshift/config.yaml <<IEOF
network:
  clusterNetwork:
  - ${CLUSTER_C_POD_CIDR}
  serviceNetwork:
  - ${CLUSTER_C_SVC_CIDR}
IEOF
EOF

    launch_vm rhel102-bootc --vmname host1
    launch_vm rhel102-bootc --vmname host2
    launch_vm rhel102-bootc --vmname host3
}

scenario_remove_vms() {
    remove_vm host1
    remove_vm host2
    remove_vm host3
}

scenario_run_tests() {
    if ! configure_c2cc_hosts; then
        return 1
    fi

    # Retrieve host2's kubeconfig
    local -r host2_ip=$(get_vm_property host2 ip)
    local -r kubeconfig_b="${SCENARIO_INFO_DIR}/${SCENARIO}/kubeconfig-b"
    
    # Retrieve host3's kubeconfig
    local -r host3_ip=$(get_vm_property host3 ip)
    local -r kubeconfig_c="${SCENARIO_INFO_DIR}/${SCENARIO}/kubeconfig-c"

    # Wait for host2 and host3 to be fully ready (run_tests only waits for host1)
    wait_for_microshift_to_be_ready host2
    wait_for_microshift_to_be_ready host3

    run_command_on_vm host2 "sudo cp /var/lib/microshift/resources/kubeadmin/${host2_ip}/kubeconfig /tmp/kubeconfig-b && sudo chmod 644 /tmp/kubeconfig-b"
    run_command_on_vm host3 "sudo cp /var/lib/microshift/resources/kubeadmin/${host3_ip}/kubeconfig /tmp/kubeconfig-c && sudo chmod 644 /tmp/kubeconfig-c"
    copy_file_from_vm host2 "/tmp/kubeconfig-b" "${kubeconfig_b}"
    copy_file_from_vm host3 "/tmp/kubeconfig-c" "${kubeconfig_c}"

    run_tests host1 \
        --variable "CLUSTER_A_POD_CIDR:${CLUSTER_A_POD_CIDR}" \
        --variable "CLUSTER_A_SVC_CIDR:${CLUSTER_A_SVC_CIDR}" \
        --variable "CLUSTER_A_DOMAIN:${CLUSTER_A_DOMAIN}" \
        --variable "CLUSTER_B_POD_CIDR:${CLUSTER_B_POD_CIDR}" \
        --variable "CLUSTER_B_SVC_CIDR:${CLUSTER_B_SVC_CIDR}" \
        --variable "CLUSTER_B_DOMAIN:${CLUSTER_B_DOMAIN}" \
        --variable "KUBECONFIG_B:${kubeconfig_b}" \
        --variable "CLUSTER_C_POD_CIDR:${CLUSTER_C_POD_CIDR}" \
        --variable "CLUSTER_C_SVC_CIDR:${CLUSTER_C_SVC_CIDR}" \
        --variable "CLUSTER_C_DOMAIN:${CLUSTER_C_DOMAIN}" \
        --variable "KUBECONFIG_C:${kubeconfig_c}" \
        suites/c2cc/
}
