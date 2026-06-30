#!/bin/bash
# shellcheck source=test/scenarios-rpm/common-scenarios-rpm.sh
source "${TESTDIR}/scenarios-rpm/common-scenarios-rpm.sh"

scenario_create_vms() {
    prepare_kickstart host1 kickstart-liveimg.ks.template ""
    launch_vm "${RPM_INSTALLER_IMAGE}" --vm_vcpus 6
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_setup_vms() {
    rpm_configure_vm
    rpm_install_microshift

    # Install low-latency RPM — tuned needs a reboot to activate the profile
    run_command_on_vm host1 "sudo dnf install -y microshift-low-latency"
    rpm_reboot_and_wait
}

scenario_run_tests() {
    run_tests host1 \
        --exitonfailure \
        suites/tuned/profile.robot \
        suites/tuned/microshift-tuned.robot \
        suites/tuned/workload-partitioning.robot \
        suites/tuned/uncore-cache.robot
}
