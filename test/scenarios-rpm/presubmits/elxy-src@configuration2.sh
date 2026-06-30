#!/bin/bash
# shellcheck source=test/scenarios-rpm/common-scenarios-rpm.sh
source "${TESTDIR}/scenarios-rpm/common-scenarios-rpm.sh"

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
        suites/configuration2/
}
