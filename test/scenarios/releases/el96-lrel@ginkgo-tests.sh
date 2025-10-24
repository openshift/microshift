#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

start_image="rhel-9.6-microshift-brew-optionals-4.${MINOR_VERSION}-${LATEST_RELEASE_TYPE}"

scenario_create_vms() {
    if ! does_commit_exist "${start_image}"; then
        echo "Image '${start_image}' not found - skipping test"
        return 0
    fi

    prepare_kickstart host1 kickstart.ks.template "${start_image}"
    launch_vm --vm_disksize 30
}

scenario_remove_vms() {
    if ! does_commit_exist "${start_image}"; then
        echo "Image '${start_image}' not found - skipping test"
        return 0
    fi

    remove_vm host1
}

scenario_run_tests() {
    if ! does_commit_exist "${start_image}"; then
        echo "Image '${start_image}' not found - skipping test"
        return 0
    fi
    
    # Wait for MicroShift to be ready
    wait_for_microshift_to_be_ready host1

    run_gingko_tests host1 "~Disruptive"
}
