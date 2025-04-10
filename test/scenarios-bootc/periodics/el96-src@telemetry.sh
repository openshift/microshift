#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

VM_BRIDGE_IP="$(get_vm_bridge_ip default)"

scenario_create_vms() {
    prepare_kickstart host1 kickstart-bootc.ks.template rhel96-bootc-source
    launch_vm --boot_blueprint rhel96-bootc
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    run_tests host1 \
        --variable "PROXY_HOST:${VM_BRIDGE_IP}" \
        --variable "PROXY_PORT:9001" \
        --variable "PROMETHEUS_HOST:$(hostname)" \
        --variable "PROMETHEUS_PORT:9092" \
        suites/telemetry/telemetry.robot
}
