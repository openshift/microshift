#!/bin/bash

# Sourced from scenario.sh and uses functions defined there. 

scenario_create_vms() {
    [[ "${UNAME_M}" =~ aarch64 ]] && { record_junit "setup" "scenario_create_vms" "SKIPPED"; exit 0; }

    prepare_kickstart host1 kickstart.ks.template rhel-9.6-microshift-source-isolated true
    launch_vm --boot_blueprint rhel-9.6-microshift-source-isolated --fips
}

scenario_remove_vms() {
    [[ "${UNAME_M}" =~ aarch64 ]] && { echo "Only x86_64 architecture is supported with FIPS";  exit 0; }

    remove_vm host1
}

scenario_run_tests() {
    [[ "${UNAME_M}" =~ aarch64 ]] && { echo "Only x86_64 architecture is supported with FIPS"; exit 0; }

    run_tests host1 suites/fips/
}
