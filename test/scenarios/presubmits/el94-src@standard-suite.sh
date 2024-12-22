#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart.ks.template rhel-9.4-microshift-source
    prepare_kickstart host2 kickstart.ks.template rhel-9.4-microshift-source
    prepare_kickstart host3 kickstart.ks.template rhel-9.4-microshift-source

    launch_vm --vmname host1 &
    launch_vm --vmname host1 &
    launch_vm --vmname host1 &
    wait
}

scenario_remove_vms() {
    remove_vm host1
    remove_vm host2
    remove_vm host3
}

scenario_run_tests() {
    run_tests host1 suites/standard/group1 suites/selinux/validate-selinux-policy.robot
    run_tests host2 suites/standard/group2/
    run_tests host3 suites/standard/group3/
}
