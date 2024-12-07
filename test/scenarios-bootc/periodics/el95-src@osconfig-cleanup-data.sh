#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart-bootc.ks.template rhel95-bootc-source
    launch_vm --boot_blueprint rhel95-bootc --bootc
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    run_tests host1 suites/osconfig/cleanup-data.robot
}
