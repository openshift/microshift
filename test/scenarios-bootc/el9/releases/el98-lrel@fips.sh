#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

start_image="rhel98-bootc-brew-lrel-fips"
SCENARIO_VCPUS=4
SCENARIO_FIPS=true

# Opt-in to dynamic VM scheduling by declaring requirements
dynamic_schedule_requirements() {
    echo "boot_image=${start_image}"
}

check_platform() {
    if [[ "${UNAME_M}" =~ aarch64 ]] ; then
        record_junit "setup" "scenario_create_vms" "SKIPPED"
        exit 0
    fi
}

scenario_create_vms() {
    check_platform
    exit_if_image_not_found "${start_image}"

    prepare_kickstart host1 kickstart-bootc.ks.template "${start_image}"
    launch_vm rhel98-bootc
}

scenario_remove_vms() {
    check_platform
    exit_if_image_not_found "${start_image}"

    remove_vm host1
}

scenario_run_tests() {
    check_platform
    exit_if_image_not_found "${start_image}"

    run_tests host1 suites/fips/
}
