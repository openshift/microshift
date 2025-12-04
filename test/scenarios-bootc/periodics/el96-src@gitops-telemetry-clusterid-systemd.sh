#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart-bootc.ks.template rhel96-bootc-source-gitops
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
        suites/gitops/ \
        suites/telemetry/telemetry.robot \
        suites/osconfig/clusterid.robot \
        suites/osconfig/systemd-resolved.robot
}
