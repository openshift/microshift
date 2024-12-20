#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    # Use a layer with preloaded container images to speed up MicroShift startup
    # for this long running suite
    prepare_kickstart host1 kickstart-bootc.ks.template cos9-bootc-source-isolated
    launch_vm --boot_blueprint centos9-bootc
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    run_tests host1 suites/standard1/ suites/selinux/validate-selinux-policy.robot
}
