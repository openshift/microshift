#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart-bootc.ks.template cos9-bootc-source
    prepare_kickstart host2 kickstart-bootc.ks.template cos9-bootc-source
    prepare_kickstart host3 kickstart-bootc.ks.template cos9-bootc-source

    launch_vm --vmname host1 --boot_blueprint centos9-bootc &
    launch_vm --vmname host2 --boot_blueprint centos9-bootc &
    launch_vm --vmname host3 --boot_blueprint centos9-bootc &
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
