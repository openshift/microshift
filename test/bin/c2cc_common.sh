#!/bin/bash
#
# This script contains common functions used by C2CC scenarios.

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

# Dual-stack secondary CIDRs (empty unless overridden by scenario)
CLUSTER_A_POD_CIDR_DUAL=""
CLUSTER_A_SVC_CIDR_DUAL=""
CLUSTER_B_POD_CIDR_DUAL=""
CLUSTER_B_SVC_CIDR_DUAL=""
CLUSTER_C_POD_CIDR_DUAL=""
CLUSTER_C_SVC_CIDR_DUAL=""

export TEST_RANDOMIZATION=suites

# c2cc_setup_ipv6 overrides CIDRs and mirror registry for single-stack IPv6
# scenarios. If a network name is provided, also sets VM_BRIDGE_IP and
# WEB_SERVER_URL for that network (standard IPv6 scenarios). Jumbo IPv6
# scenarios omit the network arg and set VM_BRIDGE_IP later, after creating
# the jumbo-ipv6 network.
# shellcheck disable=SC2120
c2cc_setup_ipv6() {
    local -r network="${1:-}"

    # shellcheck disable=SC2034  # used by scenario.sh and kickstart templates
    MIRROR_REGISTRY_URL="$(hostname):${MIRROR_REGISTRY_PORT}/microshift"

    CLUSTER_A_POD_CIDR="fd01::/48"
    CLUSTER_A_SVC_CIDR="fd02::/112"
    CLUSTER_A_DOMAIN="cluster-a.remote"
    CLUSTER_B_POD_CIDR="fd04::/48"
    CLUSTER_B_SVC_CIDR="fd05::/112"
    CLUSTER_B_DOMAIN="cluster-b.remote"
    CLUSTER_C_POD_CIDR="fd07::/48"
    CLUSTER_C_SVC_CIDR="fd08::/112"
    CLUSTER_C_DOMAIN="cluster-c.remote"

    if [[ -n "${network}" ]]; then
        # shellcheck disable=SC2034  # used by scenario.sh
        VM_BRIDGE_IP="$(get_vm_bridge_ip "${network}")"
        # shellcheck disable=SC2034  # used by scenario.sh
        WEB_SERVER_URL="http://[${VM_BRIDGE_IP}]:${WEB_SERVER_PORT}"
    fi
}

get_host_ip() {
    local host=$1
    get_vm_property "${host}" ip || { echo "failed to get ${host} ip" >&2; return 1; }
}

get_host_ipv6() {
    local host=$1
    local ipv6
    # Get the global-scope IPv6 address from the VM (excluding link-local fe80::)
    ipv6=$(run_command_on_vm "${host}" \
        "ip -6 addr show scope global | grep -oP '(?<=inet6 )([0-9a-f:]+)' | head -1" \
        | tail -1 | tr -d '\r')
    if [[ -z "${ipv6}" ]]; then
        echo "failed to get ${host} IPv6 address" >&2
        return 1
    fi
    echo "${ipv6}"
}

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
    # Remaining args are sets of 7:
    #   remote_ip remote_ipv6 remote_pod_cidr remote_svc_cidr remote_domain remote_pod_cidr_dual remote_svc_cidr_dual
    # remote_ipv6 and the last two may be empty for single-stack scenarios.

    run_command_on_vm "${host}" "sudo mkdir -p /etc/microshift/config.d"

    # Build the YAML config with all remote clusters
    local yaml_content
    yaml_content="clusterToCluster:"$'\n'"  remoteClusters:"
    local firewall_cidrs=()

    while [ $# -gt 0 ]; do
        local remote_ip=$1
        local remote_ipv6=${2:-}
        local remote_pod_cidr=$3
        local remote_svc_cidr=$4
        local remote_domain=$5
        local remote_pod_cidr_dual=${6:-}
        local remote_svc_cidr_dual=${7:-}
        shift 7

        yaml_content+=$'\n'"  - nextHop:"
        yaml_content+=$'\n'"    - ${remote_ip}"
        if [ -n "${remote_ipv6}" ]; then
            yaml_content+=$'\n'"    - ${remote_ipv6}"
        fi
        yaml_content+=$'\n'"    clusterNetwork:"
        yaml_content+=$'\n'"    - ${remote_pod_cidr}"
        if [ -n "${remote_pod_cidr_dual}" ]; then
            yaml_content+=$'\n'"    - ${remote_pod_cidr_dual}"
        fi
        yaml_content+=$'\n'"    serviceNetwork:"
        yaml_content+=$'\n'"    - ${remote_svc_cidr}"
        if [ -n "${remote_svc_cidr_dual}" ]; then
            yaml_content+=$'\n'"    - ${remote_svc_cidr_dual}"
        fi
        yaml_content+=$'\n'"    domain: ${remote_domain}"

        firewall_cidrs+=("${remote_pod_cidr}" "${remote_svc_cidr}")
        if [ -n "${remote_pod_cidr_dual}" ]; then
            firewall_cidrs+=("${remote_pod_cidr_dual}")
        fi
        if [ -n "${remote_svc_cidr_dual}" ]; then
            firewall_cidrs+=("${remote_svc_cidr_dual}")
        fi
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
    host1_ip=$(get_host_ip host1) || return 1
    host2_ip=$(get_host_ip host2) || return 1
    host3_ip=$(get_host_ip host3) || return 1
    readonly host1_ip host2_ip host3_ip

    local host1_ipv6="" host2_ipv6="" host3_ipv6=""
    if [ -n "${CLUSTER_A_POD_CIDR_DUAL}" ]; then
        host1_ipv6=$(get_host_ipv6 host1) || return 1
        host2_ipv6=$(get_host_ipv6 host2) || return 1
        host3_ipv6=$(get_host_ipv6 host3) || return 1
    fi
    readonly host1_ipv6 host2_ipv6 host3_ipv6

    wait_for_greenboot_on_hosts "${pre_junit_label}"

    configure_c2cc_host host1 \
        "${host2_ip}" "${host2_ipv6}" "${CLUSTER_B_POD_CIDR}" "${CLUSTER_B_SVC_CIDR}" "${CLUSTER_B_DOMAIN}" "${CLUSTER_B_POD_CIDR_DUAL}" "${CLUSTER_B_SVC_CIDR_DUAL}" \
        "${host3_ip}" "${host3_ipv6}" "${CLUSTER_C_POD_CIDR}" "${CLUSTER_C_SVC_CIDR}" "${CLUSTER_C_DOMAIN}" "${CLUSTER_C_POD_CIDR_DUAL}" "${CLUSTER_C_SVC_CIDR_DUAL}"

    configure_c2cc_host host2 \
        "${host1_ip}" "${host1_ipv6}" "${CLUSTER_A_POD_CIDR}" "${CLUSTER_A_SVC_CIDR}" "${CLUSTER_A_DOMAIN}" "${CLUSTER_A_POD_CIDR_DUAL}" "${CLUSTER_A_SVC_CIDR_DUAL}" \
        "${host3_ip}" "${host3_ipv6}" "${CLUSTER_C_POD_CIDR}" "${CLUSTER_C_SVC_CIDR}" "${CLUSTER_C_DOMAIN}" "${CLUSTER_C_POD_CIDR_DUAL}" "${CLUSTER_C_SVC_CIDR_DUAL}"

    configure_c2cc_host host3 \
        "${host1_ip}" "${host1_ipv6}" "${CLUSTER_A_POD_CIDR}" "${CLUSTER_A_SVC_CIDR}" "${CLUSTER_A_DOMAIN}" "${CLUSTER_A_POD_CIDR_DUAL}" "${CLUSTER_A_SVC_CIDR_DUAL}" \
        "${host2_ip}" "${host2_ipv6}" "${CLUSTER_B_POD_CIDR}" "${CLUSTER_B_SVC_CIDR}" "${CLUSTER_B_DOMAIN}" "${CLUSTER_B_POD_CIDR_DUAL}" "${CLUSTER_B_SVC_CIDR_DUAL}"

    wait_for_greenboot_on_hosts "${post_junit_label}"
}

# inject_kickstart_network writes a network config block into a host's
# kickstart post-microshift.cfg.
# Args: host pod_cidr svc_cidr [pod_cidr_dual] [svc_cidr_dual]
inject_kickstart_network() {
    local -r host=$1
    local -r pod_cidr=$2
    local -r svc_cidr=$3
    local -r pod_cidr_dual=${4:-}
    local -r svc_cidr_dual=${5:-}

    local -r ks_dir="${SCENARIO_INFO_DIR}/${SCENARIO}/vms/${host}"
    local network_yaml=""
    network_yaml+="  clusterNetwork:"$'\n'
    network_yaml+="  - ${pod_cidr}"$'\n'
    if [ -n "${pod_cidr_dual}" ]; then
        network_yaml+="  - ${pod_cidr_dual}"$'\n'
    fi
    network_yaml+="  serviceNetwork:"$'\n'
    network_yaml+="  - ${svc_cidr}"$'\n'
    if [ -n "${svc_cidr_dual}" ]; then
        network_yaml+="  - ${svc_cidr_dual}"$'\n'
    fi

    cat >> "${ks_dir}/post-microshift.cfg" <<EOF
cat - >>/etc/microshift/config.yaml <<IEOF
network:
${network_yaml}IEOF
EOF
}

# inject_kickstart_mtu configures jumbo MTU on the guest VM via kickstart.
# The NIC and pod MTU must be set before MicroShift first starts — OVN-K
# bakes the MTU into its database at initial creation and does not update
# it on restart. Three mechanisms set the NIC MTU (NM conf.d, NM dispatcher,
# systemd oneshot) and ovn.yaml sets the pod MTU.
# Args: host mtu
inject_kickstart_mtu() {
    local -r host="${1}"
    local -r mtu="${2}"

    local -r ks_dir="${SCENARIO_INFO_DIR}/${SCENARIO}/vms/${host}"
    cat >> "${ks_dir}/post-microshift.cfg" <<EOF
# NM conf.d -- default ethernet MTU for auto-created connections
mkdir -p /etc/NetworkManager/conf.d
cat > /etc/NetworkManager/conf.d/99-jumbo-mtu.conf <<'NMCONF'
[connection-ethernet-jumbo]
match-device=type:ethernet
ethernet.mtu=${mtu}
NMCONF

# NM dispatcher -- runs ip link set during connection activation
mkdir -p /etc/NetworkManager/dispatcher.d
cat > /etc/NetworkManager/dispatcher.d/99-jumbo-mtu <<'DISPEOF'
#!/bin/bash
[ "\$2" != "up" ] && exit 0
ip link set dev "\$1" mtu ${mtu} 2>&1 || logger -t jumbo-mtu "Failed to set MTU ${mtu} on \$1"
DISPEOF
chmod 755 /etc/NetworkManager/dispatcher.d/99-jumbo-mtu
restorecon -R /etc/NetworkManager/dispatcher.d/ 2>&1 || logger -t jumbo-mtu "restorecon failed for dispatcher.d"

# Systemd service -- runs before microshift.service.
# Script in /etc/ because only /etc/ and /var/ persist across reboots on bootc.
cat > /etc/set-jumbo-mtu.sh <<'SCRIPTEOF'
#!/bin/bash
for dev in \$(ip -o link show type ether | awk -F: '{print \$2}' | tr -d ' '); do
    ip link set dev "\$dev" mtu ${mtu}
done
SCRIPTEOF
chmod 755 /etc/set-jumbo-mtu.sh
restorecon /etc/set-jumbo-mtu.sh 2>&1 || logger -t jumbo-mtu "restorecon failed for set-jumbo-mtu.sh"

cat > /etc/systemd/system/set-jumbo-mtu.service <<'SVCEOF'
[Unit]
Description=Set jumbo frame MTU on ethernet interfaces
Before=microshift.service
After=NetworkManager-wait-online.service

[Service]
Type=oneshot
ExecStart=/etc/set-jumbo-mtu.sh
RemainAfterExit=yes

[Install]
WantedBy=multi-user.target
SVCEOF
mkdir -p /etc/systemd/system/multi-user.target.wants
ln -sf /etc/systemd/system/set-jumbo-mtu.service /etc/systemd/system/multi-user.target.wants/set-jumbo-mtu.service

# Set pod MTU to match NIC MTU. MicroShift auto-detection fails in C2CC
# VMs (no default route), falling back to 1500 regardless of NIC MTU.
mkdir -p /etc/microshift
echo "mtu: ${mtu}" > /etc/microshift/ovn.yaml
EOF
}

c2cc_create_vms() {
    local -r boot_commit_ref="${1}"
    local -r boot_blueprint="${2}"
    local -r network="${3:-default}"
    local -r ip_family="${4:-ipv4}"
    local -r network_mtu="${5:-}"

    # Prepare kickstart for all hosts
    # prepare_kickstart args: vmname template commit_ref fips_enabled ipv6_only
    local ipv6_args=""
    [ "${ip_family}" = "ipv6" ] && ipv6_args="false true"

    for host in host1 host2 host3; do
        #unquoted expansion so it expands to nothing or two arguments
        # shellcheck disable=SC2086
        prepare_kickstart "${host}" kickstart-bootc.ks.template "${boot_commit_ref}" ${ipv6_args}
    done

    # Inject CIDRs into kickstart config so MicroShift boots with the correct
    # network from the start (no cleanup-data needed).

    # For dual-stack, host1 needs explicit CIDRs too (MicroShift defaults are IPv4-only).
    if [ -n "${CLUSTER_A_POD_CIDR_DUAL}" ]; then
        inject_kickstart_network host1 \
            "${CLUSTER_A_POD_CIDR}" "${CLUSTER_A_SVC_CIDR}" \
            "${CLUSTER_A_POD_CIDR_DUAL}" "${CLUSTER_A_SVC_CIDR_DUAL}"
    fi

    inject_kickstart_network host2 \
        "${CLUSTER_B_POD_CIDR}" "${CLUSTER_B_SVC_CIDR}" \
        "${CLUSTER_B_POD_CIDR_DUAL}" "${CLUSTER_B_SVC_CIDR_DUAL}"

    inject_kickstart_network host3 \
        "${CLUSTER_C_POD_CIDR}" "${CLUSTER_C_SVC_CIDR}" \
        "${CLUSTER_C_POD_CIDR_DUAL}" "${CLUSTER_C_SVC_CIDR_DUAL}"

    # Set the NIC MTU via NetworkManager in the kickstart so the guest boots
    # with the correct MTU before MicroShift starts.
    local mtu_args=()
    if [ -n "${network_mtu}" ]; then
        mtu_args=(--network_mtu "${network_mtu}")
        for host in host1 host2 host3; do
            inject_kickstart_mtu "${host}" "${network_mtu}"
        done
    fi

    launch_vm "${boot_blueprint}" --vmname host1 --network "${network}" "${mtu_args[@]}"
    launch_vm "${boot_blueprint}" --vmname host2 --network "${network}" "${mtu_args[@]}"
    launch_vm "${boot_blueprint}" --vmname host3 --network "${network}" "${mtu_args[@]}"
}

c2cc_remove_vms() {
    remove_vm host1
    remove_vm host2
    remove_vm host3
}

c2cc_run_tests() {
    local -r suites_dir="${1}"
    local -r foreign_cidr="${2:-}"
    local -r ip_family="${3:-}"
    local -r extra_vars="${4:-}"

    local foreign_cidr_var=""
    if [ -n "${foreign_cidr}" ]; then
        foreign_cidr_var="--variable FOREIGN_CIDR:${foreign_cidr}"
    fi

    local ip_family_var=""
    if [ -n "${ip_family}" ]; then
        ip_family_var="--variable IP_FAMILY:${ip_family}"
    fi

    local target_ref_var=""
    if [ -n "${C2CC_TARGET_REF:-}" ]; then
        target_ref_var="--variable TARGET_REF:${C2CC_TARGET_REF}"
    fi

    local host2_vm host3_vm
    host2_vm=$(full_vm_name host2) || return 1
    host3_vm=$(full_vm_name host3) || return 1

    local host1_ip host2_ip host3_ip
    host1_ip=$(get_host_ip host1) || return 1
    host2_ip=$(get_host_ip host2) || return 1
    host3_ip=$(get_host_ip host3) || return 1
    readonly host1_ip host2_ip host3_ip

    local -r kubeconfig_a="${SCENARIO_INFO_DIR}/${SCENARIO}/kubeconfig-a"
    local -r kubeconfig_b="${SCENARIO_INFO_DIR}/${SCENARIO}/kubeconfig-b"
    local -r kubeconfig_c="${SCENARIO_INFO_DIR}/${SCENARIO}/kubeconfig-c"

    # Wait for all hosts to be fully ready before fetching kubeconfigs
    wait_for_microshift_to_be_ready host1
    wait_for_microshift_to_be_ready host2
    wait_for_microshift_to_be_ready host3

    run_command_on_vm host1 "sudo cp /var/lib/microshift/resources/kubeadmin/${host1_ip}/kubeconfig /tmp/kubeconfig-a && sudo chmod 644 /tmp/kubeconfig-a"
    run_command_on_vm host2 "sudo cp /var/lib/microshift/resources/kubeadmin/${host2_ip}/kubeconfig /tmp/kubeconfig-b && sudo chmod 644 /tmp/kubeconfig-b"
    run_command_on_vm host3 "sudo cp /var/lib/microshift/resources/kubeadmin/${host3_ip}/kubeconfig /tmp/kubeconfig-c && sudo chmod 644 /tmp/kubeconfig-c"
    copy_file_from_vm host1 "/tmp/kubeconfig-a" "${kubeconfig_a}"
    copy_file_from_vm host2 "/tmp/kubeconfig-b" "${kubeconfig_b}"
    copy_file_from_vm host3 "/tmp/kubeconfig-c" "${kubeconfig_c}"

    local dual_cidr_vars=""
    if [ -n "${CLUSTER_A_POD_CIDR_DUAL}" ]; then
        local host1_ipv6 host2_ipv6 host3_ipv6
        host1_ipv6=$(get_host_ipv6 host1) || return 1
        host2_ipv6=$(get_host_ipv6 host2) || return 1
        host3_ipv6=$(get_host_ipv6 host3) || return 1
        dual_cidr_vars+=" --variable CLUSTER_A_POD_CIDR_DUAL:${CLUSTER_A_POD_CIDR_DUAL}"
        dual_cidr_vars+=" --variable CLUSTER_A_SVC_CIDR_DUAL:${CLUSTER_A_SVC_CIDR_DUAL}"
        dual_cidr_vars+=" --variable CLUSTER_B_POD_CIDR_DUAL:${CLUSTER_B_POD_CIDR_DUAL}"
        dual_cidr_vars+=" --variable CLUSTER_B_SVC_CIDR_DUAL:${CLUSTER_B_SVC_CIDR_DUAL}"
        dual_cidr_vars+=" --variable CLUSTER_C_POD_CIDR_DUAL:${CLUSTER_C_POD_CIDR_DUAL}"
        dual_cidr_vars+=" --variable CLUSTER_C_SVC_CIDR_DUAL:${CLUSTER_C_SVC_CIDR_DUAL}"
        dual_cidr_vars+=" --variable HOST1_IPV6:${host1_ipv6}"
        dual_cidr_vars+=" --variable HOST2_IPV6:${host2_ipv6}"
        dual_cidr_vars+=" --variable HOST3_IPV6:${host3_ipv6}"
    fi

    # shellcheck disable=SC2086
    run_tests host1 \
        --variable "CLUSTER_A_POD_CIDR:${CLUSTER_A_POD_CIDR}" \
        --variable "CLUSTER_A_SVC_CIDR:${CLUSTER_A_SVC_CIDR}" \
        --variable "CLUSTER_A_DOMAIN:${CLUSTER_A_DOMAIN}" \
        --variable "KUBECONFIG_A:${kubeconfig_a}" \
        --variable "CLUSTER_B_POD_CIDR:${CLUSTER_B_POD_CIDR}" \
        --variable "CLUSTER_B_SVC_CIDR:${CLUSTER_B_SVC_CIDR}" \
        --variable "CLUSTER_B_DOMAIN:${CLUSTER_B_DOMAIN}" \
        --variable "KUBECONFIG_B:${kubeconfig_b}" \
        --variable "CLUSTER_C_POD_CIDR:${CLUSTER_C_POD_CIDR}" \
        --variable "CLUSTER_C_SVC_CIDR:${CLUSTER_C_SVC_CIDR}" \
        --variable "CLUSTER_C_DOMAIN:${CLUSTER_C_DOMAIN}" \
        --variable "KUBECONFIG_C:${kubeconfig_c}" \
        --variable "HOST2_VM_NAME:${host2_vm}" \
        --variable "HOST3_VM_NAME:${host3_vm}" \
        ${foreign_cidr_var} \
        ${ip_family_var} \
        ${target_ref_var} \
        ${dual_cidr_vars} \
        ${extra_vars} \
        --variable "BOOTC_REGISTRY:${MIRROR_REGISTRY_URL}" \
        "${suites_dir}"
}

# ipsec specific functions

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
    host1_ip=$(get_host_ip host1) || return 1
    host2_ip=$(get_host_ip host2) || return 1
    host3_ip=$(get_host_ip host3) || return 1
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
