#!/bin/bash
# shellcheck source=test/bin/scenario_rpm.sh
source "${TESTDIR}/bin/scenario_rpm.sh"

scenario_create_vms() {
    prepare_kickstart host1 kickstart-liveimg.ks.template ""
    launch_vm "${RPM_INSTALLER_IMAGE}" --network "${VM_DUAL_STACK_NETWORK}"
}
scenario_setup_vms() {
    rpm_configure_vm
    rpm_install_microshift
    # Reboot to ensure clean IPv6/dual-stack network state after NM restart
    rpm_reboot_and_wait
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    run_tests host1 suites/ipv6/dualstack.robot
}
