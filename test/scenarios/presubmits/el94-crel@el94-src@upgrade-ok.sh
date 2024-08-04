#!/bin/bash

# This scenario was disabled because osbuild takes a newest cri-o possible for `cri-o >= 1.25.0` requirement,
# which happens to be crio >= 1.29 which no longer ships `crun` runtime definition.
# TODO: Re-enable when new MicroShift EC (with crun runtime definition in cri-o config) is produced.

# Sourced from cleanup_scenario.sh and uses functions defined there.

start_commit=rhel-9.4-microshift-crel

scenario_create_vms() {
    if ! does_commit_exist "${start_commit}"; then
        echo "Commit '${start_commit}' not found in ostree repo - skipping test"
        return 0
    fi
    prepare_kickstart host1 kickstart.ks.template "${start_commit}"
    launch_vm host1 rhel-9.4
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
