#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.
# shellcheck source=test/scenarios-rpm/common-scenarios-rpm.sh
source "${TESTDIR}/scenarios-rpm/common-scenarios-rpm.sh"

RPM_RHEL_MAJOR=10

scenario_create_vms() {
    prepare_kickstart host1 kickstart-liveimg.ks.template ""
    launch_vm rhel102-installer
    configure_vm_firewall host1
    subscription_manager_register host1

    # Enable the rhocp current and previous minor version repos.
    configure_rhocp_repo "${RHOCP_MINOR_Y}"       "${MAJOR_VERSION}" "${MINOR_VERSION}"
    configure_rhocp_repo "${RHOCP_MINOR_Y_BETA}"  "${MAJOR_VERSION}" "${MINOR_VERSION}"
    # Pin the RHEL release to avoid package conflicts with other minor versions.
    run_command_on_vm host1 "sudo subscription-manager release --set 10.2"
    # Enable the fast-datapath repo for MicroShift networking (OVN-Kubernetes).
    run_command_on_vm host1 "sudo subscription-manager repos --enable fast-datapath-for-rhel-9-\$(uname -m)-rpms"
    # Install the dependencies for MicroShift networking (OVN-Kubernetes).
    run_command_on_vm host1 "sudo dnf install -y NetworkManager-ovs containers-common"
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    local -r reponame=$(basename "${LOCAL_REPO}")
    local -r target_version=$(local_rpm_version)
    install_microshift "${WEB_SERVER_URL}/rpm-repos/${reponame}" "${target_version}"

    run_tests host1 \
        suites/router/
}
