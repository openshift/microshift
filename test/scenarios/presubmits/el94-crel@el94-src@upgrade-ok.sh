#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

start_commit=rhel-9.4-microshift-crel

scenario_create_vms() {
    if ! does_commit_exist "${start_commit}"; then
        echo "Commit '${start_commit}' not found in ostree repo - skipping test"
        return 0
    fi
    prepare_kickstart host1 kickstart.ks.template "${start_commit}"
    launch_vm 
}

scenario_remove_vms() {
    if ! does_commit_exist "${start_commit}"; then
        echo "Commit '${start_commit}' not found in ostree repo - skipping test"
        return 0
    fi
    remove_vm host1
}

scenario_run_tests() {
    if ! does_commit_exist "${start_commit}"; then
        echo "Commit '${start_commit}' not found in ostree repo - skipping test"
        return 0
    fi
    run_tests host1 \
              --variable "TARGET_REF:rhel-9.4-microshift-source" \
              suites/upgrade/upgrade-successful.robot
}
