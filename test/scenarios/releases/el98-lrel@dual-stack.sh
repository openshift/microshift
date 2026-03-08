#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

start_image="rhel98-brew-lrel-optional"

scenario_create_vms() {
    exit_if_commit_not_found "${start_image}"

    prepare_kickstart host1 kickstart.ks.template "${start_image}"
    launch_vm --boot_blueprint rhel-9.8 --network "${VM_DUAL_STACK_NETWORK}" --vm_vcpus 4
}

scenario_remove_vms() {
    exit_if_commit_not_found "${start_image}"

    remove_vm host1
}

scenario_run_tests() {
    exit_if_commit_not_found "${start_image}"
    run_tests host1 suites/ipv6/dualstack.robot
}
