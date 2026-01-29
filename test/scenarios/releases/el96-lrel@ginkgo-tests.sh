#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

start_image="rhel96-brew-lrel-optional"

scenario_create_vms() {
    exit_if_commit_not_found "${start_image}"

    prepare_kickstart host1 kickstart.ks.template "${start_image}"
    launch_vm --vm_disksize 30 --vm_vcpus 4
}

scenario_remove_vms() {
    exit_if_commit_not_found "${start_image}"

    remove_vm host1
}

scenario_run_tests() {
    exit_if_commit_not_found "${start_image}"

    wait_for_microshift_to_be_ready host1
    run_ginkgo_tests host1 "~Disruptive"
}
