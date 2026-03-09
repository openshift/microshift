#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

start_image="rhel98-bootc-brew-lrel-optional"

scenario_create_vms() {
    exit_if_image_not_found "${start_image}"

    prepare_kickstart host1 kickstart-bootc.ks.template "${start_image}"
    launch_vm --boot_blueprint rhel98-bootc --vm_vcpus 4
}

scenario_remove_vms() {
    exit_if_image_not_found "${start_image}"

    remove_vm host1
}

scenario_run_tests() {
    exit_if_image_not_found "${start_image}"

    run_tests host1 \
        --variable "PROXY_HOST:${VM_BRIDGE_IP}" \
        --variable "PROXY_PORT:9001" \
        --variable "PROMETHEUS_HOST:$(hostname)" \
        suites/storage/ \
        suites/telemetry/telemetry.robot
}
