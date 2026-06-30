#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.
# shellcheck source=test/scenarios-rpm/common-scenarios-rpm.sh
source "${TESTDIR}/scenarios-rpm/common-scenarios-rpm.sh"

RPM_RHEL_MAJOR=9

# Redefine network-related settings to use the dedicated IPv6 network bridge
# shellcheck disable=SC2034  # used elsewhere
VM_BRIDGE_IP="$(get_vm_bridge_ip "${VM_IPV6_NETWORK}")"
# shellcheck disable=SC2034  # used elsewhere
WEB_SERVER_URL="http://[${VM_BRIDGE_IP}]:${WEB_SERVER_PORT}"
# shellcheck disable=SC2034  # used elsewhere
MIRROR_REGISTRY_URL="${VM_BRIDGE_IP}:${MIRROR_REGISTRY_PORT}"

scenario_create_vms() {
    # Enable IPv6 single stack in kickstart
    prepare_kickstart host1 kickstart-liveimg.ks.template "" false true
    launch_vm rhel98-installer --network "${VM_IPV6_NETWORK}"
    configure_vm_firewall host1
    subscription_manager_register host1

    configure_rhocp_repo "${RHOCP_MINOR_Y}"       "${MAJOR_VERSION}" "${MINOR_VERSION}"
    configure_rhocp_repo "${RHOCP_MINOR_Y_BETA}"  "${MAJOR_VERSION}" "${MINOR_VERSION}"
    run_command_on_vm host1 "sudo subscription-manager release --set 9.8"
    run_command_on_vm host1 "sudo subscription-manager repos --enable fast-datapath-for-rhel-9-\$(uname -m)-rpms"
    run_command_on_vm host1 "sudo dnf install -y NetworkManager-ovs containers-common"
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    local -r reponame=$(basename "${LOCAL_REPO}")
    local -r target_version=$(local_rpm_version)
    install_microshift "${WEB_SERVER_URL}/rpm-repos/${reponame}" "${target_version}"

    run_tests host1 suites/ipv6/singlestack.robot
}
