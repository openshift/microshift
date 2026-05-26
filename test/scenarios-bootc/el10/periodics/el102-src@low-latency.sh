#!/bin/bash

export SKIP_GREENBOOT=true
export TEST_RANDOMIZATION=none

# Sourced from scenario.sh and uses functions defined there.

start_image="rhel102-bootc-source-tuned"

scenario_create_vms() {
    exit_if_image_not_found "${start_image}"

    prepare_kickstart host1 kickstart-bootc.ks.template "${start_image}" true
    launch_vm rhel102-bootc --vm_vcpus 6
}

scenario_remove_vms() {
    exit_if_image_not_found "${start_image}"

    remove_vm host1
}

scenario_run_tests() {
    exit_if_image_not_found "${start_image}"

    # Wait for microshift-tuned to finish its initial setup, which may
    # include reboots to activate the TuneD profile. Polling the service
    # SubState is more robust than counting boots — other system events
    # (SELinux relabeling, cloud-init) can also trigger reboots.
    local -r start_time=$(date +%s)
    while true; do
        tuned_state=$(run_command_on_vm host1 \
            "sudo systemctl show -p SubState --value microshift-tuned.service" || true)
        tuned_state="${tuned_state%$'\r'*}"
        if [[ "${tuned_state}" == "dead" || "${tuned_state}" == "exited" || "${tuned_state}" == "failed" ]]; then
            break
        fi
        if [ $(( $(date +%s) - start_time )) -gt 300 ]; then
            echo "Timed out waiting for microshift-tuned to settle (state: ${tuned_state})"
            exit 1
        fi
        sleep 10
    done

    # --exitonfailure because tests within suites are meant to be ordered,
    # so don't advance to next test if current failed.

    run_tests host1 \
        --exitonfailure \
        suites/tuned/microshift-tuned.robot \
        suites/tuned/workload-partitioning.robot \
        suites/tuned/uncore-cache.robot
}
