#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

export SKIP_GREENBOOT=true

scenario_create_vms() {
    prepare_kickstart host1 kickstart-liveimg.ks.template ""
    launch_vm rhel102-installer
    configure_rpm_scenario host1 "10.2"
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    run_tests host1 suites/standard2/
}
