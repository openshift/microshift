#!/bin/bash

# Sourced from scenario.sh and uses functions defined there. 

scenario_create_vms() {
    prepare_kickstart host1 kickstart-liveimg.ks.template rhel-9.2-microshift-source-isolated true
    launch_vm host1 "rhel-9.2-microshift-source-isolated" "" "" "" "" "" "1"
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    run_tests host1 suites/fips/
}
