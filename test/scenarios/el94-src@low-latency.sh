#!/bin/bash

export SKIP_GREENBOOT=true
export TEST_RANDOMIZATION=none

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart.ks.template rhel-9.4-microshift-source-tuned
    launch_vm host1 rhel-9.4 "" 6
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    # Should not be ran immediately after creating VM because of
    # microshift-tuned rebooting the node to activate the profile.
    
    # --exitonfailure because tests within suites are meant to be ordered,
    # so don't advance to next test if current failed.

    run_tests host1 \
        --exitonfailure \
        suites/tuned/profile.robot \
        suites/tuned/microshift-tuned.robot
}
