#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

target_commit=rhel-9.6-microshift-crel

scenario_create_vms() {
    exit_if_commit_not_found "${target_commit}"

    prepare_kickstart host1 kickstart.ks.template "rhel-9.6-microshift-4.${PREVIOUS_MINOR_VERSION}"
    launch_vm
}

scenario_remove_vms() {
    exit_if_commit_not_found "${target_commit}"

    remove_vm host1
}

scenario_run_tests() {
    exit_if_commit_not_found "${target_commit}"

    run_tests host1 \
              --variable "TARGET_REF:${target_commit}" \
              suites/upgrade/upgrade-successful.robot
}
