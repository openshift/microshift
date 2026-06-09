#!/bin/bash
#
# This script contains common functions used by C2CC scenarios.

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
    local host=$1
    shift
    # Remaining args are sets of 4: remote_ip remote_pod_cidr remote_svc_cidr remote_domain (repeat)

    run_command_on_vm "${host}" "sudo mkdir -p /etc/microshift/config.d"

    # Build the YAML config with all remote clusters
    local yaml_content
    yaml_content="clusterToCluster:"$'\n'"  remoteClusters:"
    local firewall_cidrs=()

    while [ $# -gt 0 ]; do
        local remote_ip=$1
        local remote_pod_cidr=$2
        local remote_svc_cidr=$3
        local remote_domain=$4
        shift 4

        yaml_content+=$'\n'"  - nextHop: ${remote_ip}"
        yaml_content+=$'\n'"    clusterNetwork:"
        yaml_content+=$'\n'"    - ${remote_pod_cidr}"
        yaml_content+=$'\n'"    serviceNetwork:"
        yaml_content+=$'\n'"    - ${remote_svc_cidr}"
        yaml_content+=$'\n'"    domain: ${remote_domain}"

        firewall_cidrs+=("${remote_pod_cidr}" "${remote_svc_cidr}")
    done

    run_command_on_vm "${host}" "sudo tee /etc/microshift/config.d/50-c2cc.yaml > /dev/null <<EOF
${yaml_content}
EOF"

    configure_vm_firewall "${host}"
    for cidr in "${firewall_cidrs[@]}"; do
        run_command_on_vm "${host}" "sudo firewall-cmd --permanent --zone=trusted --add-source=${cidr}"
    done
    run_command_on_vm "${host}" "sudo firewall-cmd --reload"
    run_command_on_vm "${host}" "sudo systemctl restart microshift"
}

configure_c2cc_hosts() {
    local -r pre_junit_label="${1:-c2cc_pre_greenboot}"
    local -r post_junit_label="${2:-c2cc_greenboot}"

    local host1_ip host2_ip host3_ip
    host1_ip=$(get_vm_property host1 ip) || { echo "failed to get host1 ip" >&2; return 1; }
    host2_ip=$(get_vm_property host2 ip) || { echo "failed to get host2 ip" >&2; return 1; }
    host3_ip=$(get_vm_property host3 ip) || { echo "failed to get host3 ip" >&2; return 1; }
    readonly host1_ip host2_ip host3_ip

    wait_for_greenboot_on_hosts "${pre_junit_label}"

    configure_c2cc_host host1 \
        "${host2_ip}" "${CLUSTER_B_POD_CIDR}" "${CLUSTER_B_SVC_CIDR}" "${CLUSTER_B_DOMAIN}" \
        "${host3_ip}" "${CLUSTER_C_POD_CIDR}" "${CLUSTER_C_SVC_CIDR}" "${CLUSTER_C_DOMAIN}"

    configure_c2cc_host host2 \
        "${host1_ip}" "${CLUSTER_A_POD_CIDR}" "${CLUSTER_A_SVC_CIDR}" "${CLUSTER_A_DOMAIN}" \
        "${host3_ip}" "${CLUSTER_C_POD_CIDR}" "${CLUSTER_C_SVC_CIDR}" "${CLUSTER_C_DOMAIN}"

    configure_c2cc_host host3 \
        "${host1_ip}" "${CLUSTER_A_POD_CIDR}" "${CLUSTER_A_SVC_CIDR}" "${CLUSTER_A_DOMAIN}" \
        "${host2_ip}" "${CLUSTER_B_POD_CIDR}" "${CLUSTER_B_SVC_CIDR}" "${CLUSTER_B_DOMAIN}"

    wait_for_greenboot_on_hosts "${post_junit_label}"
}

c2cc_create_vms() {
    local -r boot_commit_ref="${1}"
    local -r boot_blueprint="${2}"
    local -r network="${3:-default}"
    local -r ip_family="${4:-ipv4}" 

    # Prepare kickstart for all hosts
    local ipv6_args=""
    [ "${ip_family}" = "ipv6" ] && ipv6_args="false true"

    for host in host1 host2 host3; do
        #unquoted expansion so it expands to nothing or two arguments
        # shellcheck disable=SC2086
        prepare_kickstart "${host}" kickstart-bootc.ks.template "${boot_commit_ref}" ${ipv6_args}
    done

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

    launch_vm "${boot_blueprint}" --vmname host1 --network "${network}"
    launch_vm "${boot_blueprint}" --vmname host2 --network "${network}"
    launch_vm "${boot_blueprint}" --vmname host3 --network "${network}"
}

c2cc_remove_vms() {
    remove_vm host1
    remove_vm host2
    remove_vm host3
}

c2cc_run_tests() {
    local -r suites_dir="${1}"
    local -r foreign_cidr="${2}"
    local -r ip_family="${3}"

    local foreign_cidr_var=""
    if [ -n "${foreign_cidr}" ]; then
        foreign_cidr_var="--variable FOREIGN_CIDR:${foreign_cidr}"
    fi

    local ip_family_var=""
    if [ -n "${ip_family}" ]; then
        ip_family_var="--variable IP_FAMILY:${ip_family}"
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

    # shellcheck disable=SC2086
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
        ${foreign_cidr_var} \
        ${ip_family_var} \
        "${suites_dir}"
}
