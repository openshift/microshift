#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart.ks.template el92-src
    launch_vm host1
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    run_tests host1 \
        --variable "FAKE_NEXT_MINOR_REF:el92-src-fake-y1" \
        suites-ostree/rpm-ostree-rollback.robot
}
