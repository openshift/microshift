#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

start_image="rhel98-brew-lrel-optional"

scenario_create_vms() {
    exit_if_commit_not_found "${start_image}"

    prepare_kickstart host1 kickstart.ks.template "${start_image}"
    launch_vm rhel-9.8 --vm_vcpus 4
}

scenario_remove_vms() {
    exit_if_commit_not_found "${start_image}"

    remove_vm host1
}

scenario_run_tests() {
    exit_if_commit_not_found "${start_image}"

    DEST_DIR="${RF_VENV}" "${ROOTDIR}/scripts/fetch_tools.sh" omc omg || {
        record_junit "host1" "support_tools_installed" "FAILED"
        exit 1
    }
    record_junit "host1" "support_tools_installed" "OK"

    run_tests host1 \
        suites/otp-workloads/sos-report-plugins.robot \
        suites/otp-workloads/sos-report-support-tools.robot \
        suites/otp-workloads/kcm-flags.robot
}
