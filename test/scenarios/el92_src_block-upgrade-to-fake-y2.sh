#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart.ks.template rhel-9.2-microshift-source
    launch_vm host1
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    run_tests host1 \
        --variable "TOO_NEW_MICROSHIFT_REF:rhel-9.2-microshift-source-fake-yplus2-minor" \
        suites-ostree/upgrade-block-2-minor.robot
}
