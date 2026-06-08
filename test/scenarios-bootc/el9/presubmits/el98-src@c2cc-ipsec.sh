#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.
#
# Sets up 3 MicroShift clusters with C2CC and Libreswan IPsec (tunnel mode
# protecting pod/service CIDRs).  Each host maintains IPsec tunnels forming a
# full mesh.  Tests validate ESP encapsulation, connectivity, policy
# enforcement, plaintext rejection, and MTU behaviour.

# IPsec tests have ordering dependencies (setup verification must pass before
# enforcement tests), so disable randomization.
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
    local host=$1
    shift
    # Remaining args are sets of 4: remote_ip remote_pod_cidr remote_svc_cidr remote_domain

    run_command_on_vm "${host}" "sudo mkdir -p /etc/microshift/config.d"

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
    local host1_ip host2_ip host3_ip
    host1_ip=$(get_vm_property host1 ip) || { echo "failed to get host1 ip" >&2; return 1; }
    host2_ip=$(get_vm_property host2 ip) || { echo "failed to get host2 ip" >&2; return 1; }
    host3_ip=$(get_vm_property host3 ip) || { echo "failed to get host3 ip" >&2; return 1; }
    readonly host1_ip host2_ip host3_ip

    wait_for_greenboot_on_hosts "c2cc_ipsec_pre_greenboot"

    configure_c2cc_host host1 \
        "${host2_ip}" "${CLUSTER_B_POD_CIDR}" "${CLUSTER_B_SVC_CIDR}" "${CLUSTER_B_DOMAIN}" \
        "${host3_ip}" "${CLUSTER_C_POD_CIDR}" "${CLUSTER_C_SVC_CIDR}" "${CLUSTER_C_DOMAIN}"

    configure_c2cc_host host2 \
        "${host1_ip}" "${CLUSTER_A_POD_CIDR}" "${CLUSTER_A_SVC_CIDR}" "${CLUSTER_A_DOMAIN}" \
        "${host3_ip}" "${CLUSTER_C_POD_CIDR}" "${CLUSTER_C_SVC_CIDR}" "${CLUSTER_C_DOMAIN}"

    configure_c2cc_host host3 \
        "${host1_ip}" "${CLUSTER_A_POD_CIDR}" "${CLUSTER_A_SVC_CIDR}" "${CLUSTER_A_DOMAIN}" \
        "${host2_ip}" "${CLUSTER_B_POD_CIDR}" "${CLUSTER_B_SVC_CIDR}" "${CLUSTER_B_DOMAIN}"

    wait_for_greenboot_on_hosts "c2cc_ipsec_greenboot"
}

# configure_ipsec_host writes the PSK and connection configs, initializes the
# NSS database, and starts the ipsec service on a single host.
# Libreswan, tcpdump, and firewall rules are pre-installed in the
# rhel98-bootc-source-ipsec container image.
#
# Uses tunnel mode with subnet selectors to protect C2CC traffic (pod/service
# CIDRs). MicroShift C2CC routes cross-cluster traffic as raw IP via the
# host's physical interface — there is no Geneve tunnel between hosts.
#
# Arguments:
#   $1         — VM name (host1, host2, host3)
#   $2         — this host's IP
#   $3         — local pod CIDR
#   $4         — local service CIDR
#   $5         — pre-shared key (hex string)
#   $6..N      — sets of 4: remote_ip remote_name remote_pod_cidr remote_svc_cidr
configure_ipsec_host() {
    local -r host=$1
    local -r host_ip=$2
    local -r local_pod_cidr=$3
    local -r local_svc_cidr=$4
    local -r psk=$5
    shift 5

    local secrets_content=""
    local conn_content=""
    while [ $# -gt 0 ]; do
        local remote_ip=$1
        local remote_name=$2
        local remote_pod_cidr=$3
        local remote_svc_cidr=$4
        shift 4

        secrets_content+="${host_ip} ${remote_ip} : PSK \"${psk}\""$'\n'

        conn_content+="conn c2cc-to-${remote_name}"$'\n'
        conn_content+="    type=tunnel"$'\n'
        conn_content+="    authby=secret"$'\n'
        conn_content+="    left=${host_ip}"$'\n'
        conn_content+="    right=${remote_ip}"$'\n'
        conn_content+="    leftsubnets={${local_pod_cidr} ${local_svc_cidr}}"$'\n'
        conn_content+="    rightsubnets={${remote_pod_cidr} ${remote_svc_cidr}}"$'\n'
        conn_content+="    auto=start"$'\n'
        conn_content+="    ike=aes256-sha2_256-modp2048"$'\n'
        conn_content+="    esp=aes256-sha2_256"$'\n'
        conn_content+="    failureshunt=drop"$'\n'
        conn_content+="    negotiationshunt=drop"$'\n'
        conn_content+="    ikev2=insist"$'\n'
        conn_content+=$'\n'
    done

    run_command_on_vm "${host}" "sudo tee /etc/ipsec.d/c2cc.secrets > /dev/null <<EOF
${secrets_content}EOF"
    run_command_on_vm "${host}" "sudo chmod 600 /etc/ipsec.d/c2cc.secrets"
    run_command_on_vm "${host}" "sudo restorecon -v /etc/ipsec.d/c2cc.secrets"

    run_command_on_vm "${host}" "sudo tee /etc/ipsec.d/c2cc-tunnel.conf > /dev/null <<EOF
${conn_content}EOF"

    run_command_on_vm "${host}" "sudo ipsec checknss"
    run_command_on_vm "${host}" "sudo systemctl restart ipsec"
}

wait_for_ipsec_tunnels() {
    local host=$1
    local expected_count=$2

    local attempts=0
    while [ "${attempts}" -lt 30 ]; do
        local count
        count=$(run_command_on_vm "${host}" "sudo ipsec trafficstatus 2>/dev/null | grep -c 'type=ESP' || true")
        count=$(echo "${count}" | tail -1 | tr -d '\r')
        if [ "${count}" -ge "${expected_count}" ]; then
            record_junit "${host}" "ipsec_tunnels" "OK"
            return 0
        fi
        sleep 2
        attempts=$((attempts + 1))
    done
    record_junit "${host}" "ipsec_tunnels" "FAILED"
    return 1
}

configure_ipsec() {
    local host1_ip host2_ip host3_ip
    host1_ip=$(get_vm_property host1 ip) || { echo "failed to get host1 ip" >&2; return 1; }
    host2_ip=$(get_vm_property host2 ip) || { echo "failed to get host2 ip" >&2; return 1; }
    host3_ip=$(get_vm_property host3 ip) || { echo "failed to get host3 ip" >&2; return 1; }
    readonly host1_ip host2_ip host3_ip

    local psk
    psk=$(openssl rand -hex 32) || { echo "failed to generate PSK" >&2; return 1; }
    readonly psk

    configure_ipsec_host host1 "${host1_ip}" "${CLUSTER_A_POD_CIDR}" "${CLUSTER_A_SVC_CIDR}" "${psk}" \
        "${host2_ip}" host2 "${CLUSTER_B_POD_CIDR}" "${CLUSTER_B_SVC_CIDR}" \
        "${host3_ip}" host3 "${CLUSTER_C_POD_CIDR}" "${CLUSTER_C_SVC_CIDR}"

    configure_ipsec_host host2 "${host2_ip}" "${CLUSTER_B_POD_CIDR}" "${CLUSTER_B_SVC_CIDR}" "${psk}" \
        "${host1_ip}" host1 "${CLUSTER_A_POD_CIDR}" "${CLUSTER_A_SVC_CIDR}" \
        "${host3_ip}" host3 "${CLUSTER_C_POD_CIDR}" "${CLUSTER_C_SVC_CIDR}"

    configure_ipsec_host host3 "${host3_ip}" "${CLUSTER_C_POD_CIDR}" "${CLUSTER_C_SVC_CIDR}" "${psk}" \
        "${host1_ip}" host1 "${CLUSTER_A_POD_CIDR}" "${CLUSTER_A_SVC_CIDR}" \
        "${host2_ip}" host2 "${CLUSTER_B_POD_CIDR}" "${CLUSTER_B_SVC_CIDR}"

    # Each host has 2 remote hosts × 4 subnet pairs (2 local × 2 remote CIDRs) = 8 child SAs.
    for host in host1 host2 host3; do
        if ! wait_for_ipsec_tunnels "${host}" 8; then
            return 1
        fi
    done
}

scenario_create_vms() {
    prepare_kickstart host1 kickstart-bootc.ks.template rhel98-bootc-source-ipsec
    prepare_kickstart host2 kickstart-bootc.ks.template rhel98-bootc-source-ipsec
    prepare_kickstart host3 kickstart-bootc.ks.template rhel98-bootc-source-ipsec

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

    launch_vm rhel98-bootc --vmname host1
    launch_vm rhel98-bootc --vmname host2
    launch_vm rhel98-bootc --vmname host3
}

scenario_remove_vms() {
    remove_vm host1
    remove_vm host2
    remove_vm host3
}

scenario_run_tests() {
    configure_c2cc_hosts
    configure_ipsec

    local -r host2_ip=$(get_vm_property host2 ip)
    local -r kubeconfig_b="${SCENARIO_INFO_DIR}/${SCENARIO}/kubeconfig-b"

    local -r host3_ip=$(get_vm_property host3 ip)
    local -r kubeconfig_c="${SCENARIO_INFO_DIR}/${SCENARIO}/kubeconfig-c"

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
        suites/c2cc-ipsec/
}
