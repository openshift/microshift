#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart-bootc.ks.template cos9-bootc-source
    launch_vm --boot_blueprint centos9-bootc
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    run_tests host1 suites/standard1/ suites/selinux/validate-selinux-policy.robot
}
