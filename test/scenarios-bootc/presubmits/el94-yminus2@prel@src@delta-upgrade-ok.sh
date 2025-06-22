#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart-bootc.ks.template rhel94-bootc-yminus2
    launch_vm --boot_blueprint rhel94-bootc
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    prepare_static_delta rhel94-bootc-yminus2 rhel96-bootc-prel
    consume_static_delta rhel94-bootc-yminus2 rhel96-bootc-prel

    prepare_static_delta rhel96-bootc-prel rhel96-bootc-source
    consume_static_delta rhel96-bootc-prel rhel96-bootc-source
    
    for ref in rhel96-bootc-prel-patched rhel96-bootc-source-patched ; do
        run_tests host1 \
            --variable "TARGET_REF:${ref}" \
            --variable "BOOTC_REGISTRY:${MIRROR_REGISTRY_URL}" \
            suites/upgrade/upgrade-successful.robot
    done
}
