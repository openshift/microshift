#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

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

    # Increased disk size because of the additional embedded images (especially OVMS which is ~3.5GiB)
    LVM_SYSROOT_SIZE=20480 prepare_kickstart host1 kickstart-bootc.ks.template rhel96-bootc-source-optionals
    launch_vm --boot_blueprint rhel96-bootc --vm_disksize 30
}

scenario_remove_vms() {
    check_platform

    remove_vm host1
}

scenario_run_tests() {
    check_platform

    run_tests host1 \
        suites/ai-model-serving/ai-model-serving-online.robot
}
