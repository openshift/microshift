#!/bin/bash

scenario_create_vms() {
    prepare_kickstart host1 kickstart.ks.template rhel-9.2-microshift-latest-ec
    launch_vm host1
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    run_tests host1 \
        --variable "TARGET_REF:rhel-9.2-microshift-source" \
        suites-ostree/healthy-upgrade.robot
}
