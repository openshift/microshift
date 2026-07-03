#!/bin/bash
# shellcheck source=test/bin/scenario_rpm.sh
source "${TESTDIR}/bin/scenario_rpm.sh"

scenario_create_vms() {
    prepare_kickstart host1 kickstart-liveimg.ks.template ""
    launch_vm "${RPM_INSTALLER_IMAGE}"
}
scenario_setup_vms() {
    rpm_configure_vm
    rpm_install_microshift
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    run_tests host1 \
        suites/configuration1/dns-resource-configuration.robot \
        suites/configuration2/drop-in-config.robot
}
