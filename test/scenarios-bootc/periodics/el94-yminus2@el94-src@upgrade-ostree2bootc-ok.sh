#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    # The y-1 ostree image will be fetched from the cache as it is not built
    # as part of the bootc image build procedure
    prepare_kickstart host1 kickstart.ks.template "rhel-9.4-microshift-4.${YMINUS2_MINOR_VERSION}"
    launch_vm 
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    run_tests host1 \
        --variable "TARGET_REF:rhel94-bootc-source-ostree-parent-yminus2" \
        --variable "BOOTC_REGISTRY:${BOOTC_REGISTRY_URL}" \
        suites/upgrade/upgrade-successful.robot
}
