#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart.ks.template el92-prev
    launch_vm host1
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    run_tests host1 \
        --variable "TOO_NEW_MICROSHIFT_REF:el92-src-fake-y1" \
        suites-ostree/upgrade-block-2-minor.robot
}
