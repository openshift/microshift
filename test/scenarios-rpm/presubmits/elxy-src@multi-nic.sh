#!/bin/bash
# shellcheck source=test/scenarios-rpm/common-scenarios-rpm.sh
source "${TESTDIR}/scenarios-rpm/common-scenarios-rpm.sh"

scenario_create_vms() {
    prepare_kickstart host1 kickstart-liveimg.ks.template ""
    launch_vm "${RPM_INSTALLER_IMAGE}" --network default,"${VM_MULTUS_NETWORK}"
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_setup_vms() {
    rpm_setup_and_install_microshift

    # greenboot-healthcheck is installed as a dependency but never ran (boot-time oneshot).
    # The RF test's Setup waits for it to be in "exited" state, so start it explicitly.
    run_command_on_vm host1 "sudo systemctl start greenboot-healthcheck.service"
}

scenario_run_tests() {
    local -r vmname=$(full_vm_name host1)
    local -r vm_ip1=$("${ROOTDIR}/scripts/devenv-builder/manage-vm.sh" ip -n "${vmname}" | head -1)
    local -r vm_ip2=$("${ROOTDIR}/scripts/devenv-builder/manage-vm.sh" ip -n "${vmname}" | tail -1)

    run_tests host1 \
        --variable "USHIFT_HOST_IP1:${vm_ip1}" \
        --variable "USHIFT_HOST_IP2:${vm_ip2}" \
        suites/network/multi-nic.robot
}
