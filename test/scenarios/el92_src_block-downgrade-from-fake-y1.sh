#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart.ks.template el92-src-fake-y1
    launch_vm host1
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    run_tests host1 \
        --variable "OLDER_MICROSHIFT_REF:el92-src" \
        suites-ostree/downgrade-block.robot
}
