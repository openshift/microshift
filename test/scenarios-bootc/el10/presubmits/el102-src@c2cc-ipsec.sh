#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.
#
# Sets up 3 MicroShift clusters with C2CC and Libreswan IPsec (tunnel mode
# protecting pod/service CIDRs).  Each host maintains IPsec tunnels forming a
# full mesh.  Tests validate ESP encapsulation, connectivity, policy
# enforcement, plaintext rejection, and MTU behaviour.

# shellcheck source=test/bin/c2cc_common.sh
source "${SCRIPTDIR}/c2cc_common.sh"

# IPsec tests have ordering dependencies (setup verification must pass before
# enforcement tests), so disable randomization.
export TEST_RANDOMIZATION=none

# configure_ipsec_host writes the PSK and connection configs, initializes the
# NSS database, and starts the ipsec service on a single host.
# Libreswan, tcpdump, and firewall rules are pre-installed in the
# rhel102-bootc-source-ipsec container image.
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
    c2cc_create_vms "rhel102-bootc-source-ipsec" "rhel102-bootc"
}

scenario_remove_vms() {
    c2cc_remove_vms
}

scenario_run_tests() {
    configure_c2cc_hosts "c2cc_ipsec_pre_greenboot" "c2cc_ipsec_greenboot"
    configure_ipsec

    c2cc_run_tests "suites/c2cc-ipsec/"
}
