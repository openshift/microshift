#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.
# shellcheck source=test/scenarios-rpm/common-scenarios-rpm.sh
source "${TESTDIR}/scenarios-rpm/common-scenarios-rpm.sh"

RPM_RHEL_MAJOR=9

scenario_create_vms() {
    prepare_kickstart host1 kickstart-liveimg.ks.template ""
    # Using multus as secondary network to have 2 nics in different networks.
    launch_vm rhel98-installer --network default,"${VM_MULTUS_NETWORK}"
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

    # greenboot-healthcheck is installed as a dependency but never ran (boot-time oneshot).
    # The RF test's Setup waits for it to be in "exited" state, so start it explicitly.
    run_command_on_vm host1 "sudo systemctl start greenboot-healthcheck.service"

    local -r vmname=$(full_vm_name host1)
    local -r vm_ip1=$("${ROOTDIR}/scripts/devenv-builder/manage-vm.sh" ip -n "${vmname}" | head -1)
    local -r vm_ip2=$("${ROOTDIR}/scripts/devenv-builder/manage-vm.sh" ip -n "${vmname}" | tail -1)

    run_tests host1 \
        --variable "USHIFT_HOST_IP1:${vm_ip1}" \
        --variable "USHIFT_HOST_IP2:${vm_ip2}" \
        suites/network/multi-nic.robot
}
