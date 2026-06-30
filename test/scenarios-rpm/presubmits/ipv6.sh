#!/bin/bash
# shellcheck source=test/scenarios-rpm/common-scenarios-rpm.sh
source "${TESTDIR}/scenarios-rpm/common-scenarios-rpm.sh"

# Redefine network-related settings to use the dedicated IPv6 network bridge
# shellcheck disable=SC2034  # used elsewhere
VM_BRIDGE_IP="$(get_vm_bridge_ip "${VM_IPV6_NETWORK}")"
# shellcheck disable=SC2034  # used elsewhere
WEB_SERVER_URL="http://[${VM_BRIDGE_IP}]:${WEB_SERVER_PORT}"
# shellcheck disable=SC2034  # used elsewhere
MIRROR_REGISTRY_URL="${VM_BRIDGE_IP}:${MIRROR_REGISTRY_PORT}"

scenario_create_vms() {
    prepare_kickstart host1 kickstart-liveimg.ks.template "" false true
    launch_vm "${RPM_INSTALLER_IMAGE}" --network "${VM_IPV6_NETWORK}"
    configure_vm_firewall host1
    subscription_manager_register host1
    configure_rpm_repos
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    local -r reponame=$(basename "${LOCAL_REPO}")
    install_microshift "${WEB_SERVER_URL}/rpm-repos/${reponame}" "$(local_rpm_version)"

    run_tests host1 suites/ipv6/singlestack.robot
}
