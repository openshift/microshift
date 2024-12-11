#!/bin/bash

# Sourced from scenario.sh and uses functions defined there. 

check_platform() {
    if [[ "${UNAME_M}" =~ aarch64 ]] ; then 
        record_junit "setup" "scenario_create_vms" "SKIPPED"
        exit 0
    fi
}

scenario_create_vms() {
    check_platform

    prepare_kickstart host1 kickstart-bootc.ks.template rhel95-bootc-source-fips true
    launch_vm --boot_blueprint rhel95-bootc --fips
}

scenario_remove_vms() {
    check_platform

    remove_vm host1
}

scenario_run_tests() {
    check_platform

    run_tests host1 suites/fips/
}
