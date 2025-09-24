#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart-bootc.ks.template "rhel96-bootc-brew-${LATEST_RELEASE_TYPE}-with-optional"
    launch_vm --boot_blueprint rhel96-bootc --vm_disksize 30
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    run_gingko_tests host1 "~Disruptive"
}
