#!/bin/bash

export SKIP_GREENBOOT=true
export TEST_RANDOMIZATION=none

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart.ks.template rhel-9.6-microshift-source-tuned
    launch_vm  --vm_vcpus 6
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    # Should not be ran immediately after creating VM because of
    # microshift-tuned rebooting the node to activate the profile.
    local -r start_time=$(date +%s)
    while true; do
        boot_num=$(run_command_on_vm host1 "sudo journalctl --list-boots --quiet | wc -l" || true)
        boot_num="${boot_num%$'\r'*}"
        if [[ "${boot_num}" -ge 2 ]]; then
            break
        fi
        if [ $(( $(date +%s) - start_time )) -gt 60 ]; then
            echo "Timed out waiting for VM having 2 boots"
            exit 1
        fi
        sleep 5
    done
    
    # --exitonfailure because tests within suites are meant to be ordered,
    # so don't advance to next test if current failed.

    run_tests host1 \
        --exitonfailure \
        suites/tuned/profile.robot \
        suites/tuned/microshift-tuned.robot \
        suites/tuned/workload-partitioning.robot
}
