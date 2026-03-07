#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart-bootc.ks.template rhel98-bootc-source-base
    launch_vm --boot_blueprint rhel98-bootc
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    prepare_static_delta rhel98-bootc-source-base rhel98-bootc-source
    apply_static_delta   rhel98-bootc-source-base rhel98-bootc-source

    prepare_static_delta rhel98-bootc-source rhel98-bootc-source-optionals
    apply_static_delta   rhel98-bootc-source rhel98-bootc-source-optionals
    
    for ref in rhel98-bootc-source-from-sdelta rhel98-bootc-source-optionals-from-sdelta ; do
        run_tests host1 \
            --variable "TARGET_REF:${ref}" \
            --variable "BOOTC_REGISTRY:${MIRROR_REGISTRY_URL}" \
            suites/upgrade/upgrade-successful.robot
    done
}
