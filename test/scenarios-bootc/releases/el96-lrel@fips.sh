#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

start_image="rhel96-bootc-brew-${LATEST_RELEASE_TYPE}-with-optional-fips"

check_platform() {
    if [[ "${UNAME_M}" =~ aarch64 ]] ; then
        record_junit "setup" "scenario_create_vms" "SKIPPED"
        exit 0
    fi
}

scenario_create_vms() {
    if ! does_commit_exist "${start_image}"; then
        echo "Image '${start_image}' not found - skipping test"
        return 0
    fi

    check_platform

    prepare_kickstart host1 kickstart-bootc.ks.template "${start_image}"
    launch_vm --boot_blueprint rhel96-bootc --fips
}

scenario_remove_vms() {
    check_platform

    remove_vm host1
}

scenario_run_tests() {
    check_platform

    run_tests host1 suites/fips/
}
