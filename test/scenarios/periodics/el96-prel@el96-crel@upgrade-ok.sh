#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

target_commit=rhel-9.6-microshift-crel

scenario_create_vms() {
    if ! does_commit_exist "${target_commit}"; then
        echo "Commit '${target_commit}' not found in ostree repo - skipping test"
        return 0
    fi

    prepare_kickstart host1 kickstart.ks.template "rhel-9.6-microshift-4.${PREVIOUS_MINOR_VERSION}"
    launch_vm
}

scenario_remove_vms() {
    if ! does_commit_exist "${target_commit}"; then
        echo "Commit '${target_commit}' not found in ostree repo - skipping test"
        return 0
    fi
    remove_vm host1
}

scenario_run_tests() {
    if ! does_commit_exist "${target_commit}"; then
        echo "Commit '${target_commit}' not found in ostree repo - skipping test"
        return 0
    fi
    run_tests host1 \
              --variable "TARGET_REF:${target_commit}" \
              suites/upgrade/upgrade-successful.robot
}
