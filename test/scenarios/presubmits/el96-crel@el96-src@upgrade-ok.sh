#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

start_commit=rhel-9.6-microshift-crel

scenario_create_vms() {
    exit_if_commit_not_found "${start_commit}"

    prepare_kickstart host1 kickstart.ks.template "${start_commit}"
    launch_vm 
}

scenario_remove_vms() {
    exit_if_commit_not_found "${start_commit}"

    remove_vm host1
}

scenario_run_tests() {
    exit_if_commit_not_found "${start_commit}"

    run_tests host1 \
              --variable "TARGET_REF:rhel-9.6-microshift-source" \
              suites/upgrade/upgrade-successful.robot
}
