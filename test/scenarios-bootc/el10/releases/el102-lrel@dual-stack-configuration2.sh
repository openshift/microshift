#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

start_image="rhel102-bootc-brew-lrel-optional"

scenario_create_vms() {
    exit_if_image_not_found "${start_image}"

    prepare_kickstart host1 kickstart-bootc.ks.template "${start_image}"
    launch_vm rhel102-bootc --network "${VM_DUAL_STACK_NETWORK}" --vm_vcpus 4
}

scenario_remove_vms() {
    exit_if_image_not_found "${start_image}"

    remove_vm host1
}

scenario_run_tests() {
    exit_if_image_not_found "${start_image}"
    run_tests host1 \
        suites/configuration2/apiserver-readiness.robot \
        suites/configuration2/audit-log.robot \
        suites/configuration2/data-dir.robot \
        suites/configuration2/drop-in-config.robot \
        suites/configuration2/kustomize-sources.robot \
        suites/configuration2/logging.robot
}
