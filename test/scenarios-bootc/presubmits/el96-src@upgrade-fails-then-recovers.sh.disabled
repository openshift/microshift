#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart-bootc.ks.template rhel96-bootc-source
    launch_vm --boot_blueprint rhel96-bootc
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    run_tests host1 \
        --variable "TARGET_REF:rhel96-bootc-source-fake-next-minor" \
        --variable "BOOTC_REGISTRY:${MIRROR_REGISTRY_URL}" \
        suites/upgrade/upgrade-fails-then-recovers.robot
}
