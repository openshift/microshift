#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

start_image="rhel98-bootc-brew-lrel-optional"

# Currently, RHOAI is only available for x86_64
check_platform() {
    local -r record_junit=${1:-false}

    if [[ "${UNAME_M}" =~ aarch64 ]]; then
        if "${record_junit}"; then
            record_junit "setup" "scenario_create_vms" "SKIPPED"
        fi
        exit 0
    fi
}

scenario_create_vms() {
    check_platform true
    exit_if_image_not_found "${start_image}"

    # Increased disk size because of the additional embedded images (especially OVMS which is ~3.5GiB)
    LVM_SYSROOT_SIZE=20480 prepare_kickstart host1 kickstart-bootc.ks.template "${start_image}"
    launch_vm --boot_blueprint rhel98-bootc --vm_disksize 30 --vm_vcpus 4
}

scenario_remove_vms() {
    check_platform
    exit_if_image_not_found "${start_image}"

    remove_vm host1
}

scenario_run_tests() {
    check_platform
    exit_if_image_not_found "${start_image}"

    run_tests host1 \
        suites/ai-model-serving/ai-model-serving-online.robot
}
