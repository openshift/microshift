#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    # Use a layer with preloaded container images to speed up MicroShift startup
    # for this long running suite
    prepare_kickstart host1 kickstart-bootc.ks.template rhel95-bootc-source-isolated
    launch_vm --boot_blueprint rhel95-bootc
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    run_tests host1 suites/standard1/
    # When SELinux is working on RHEL 9.6 bootc systems add following suite:
    # suites/selinux/validate-selinux-policy.robot
}
