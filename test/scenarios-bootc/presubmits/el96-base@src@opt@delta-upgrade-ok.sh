#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart-bootc.ks.template rhel96-bootc-source-base
    launch_vm --boot_blueprint rhel96-bootc
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    prepare_static_delta rhel96-bootc-source-base rhel96-bootc-source
    apply_static_delta   rhel96-bootc-source-base rhel96-bootc-source

    prepare_static_delta rhel96-bootc-source rhel96-bootc-source-optionals
    apply_static_delta   rhel96-bootc-source rhel96-bootc-source-optionals
    
    for ref in rhel96-bootc-source-from-sdelta rhel96-bootc-source-optionals-from-sdelta ; do
        run_tests host1 \
            --variable "TARGET_REF:${ref}" \
            --variable "BOOTC_REGISTRY:${MIRROR_REGISTRY_URL}" \
            suites/upgrade/upgrade-successful.robot
    done
}
