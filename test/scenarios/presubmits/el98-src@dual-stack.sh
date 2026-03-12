#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart.ks.template rhel-9.8-microshift-source
    launch_vm rhel-9.8 --network "${VM_DUAL_STACK_NETWORK}"
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    run_tests host1 suites/ipv6/dualstack.robot
}
