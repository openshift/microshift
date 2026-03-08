#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    # TODO: Add rhel98-bootc-source-gitops when the RPM package is updated not to use
    # non-existent calls from the greenboot functions include file
    prepare_kickstart host1 kickstart-bootc.ks.template rhel98-bootc-source
    launch_vm --boot_blueprint rhel98-bootc
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    # TODO: Add suites/gitops/ when the RPM package is updated not to use
    # non-existent calls from the greenboot functions include file
    run_tests host1 \
        --variable "PROXY_HOST:${VM_BRIDGE_IP}" \
        --variable "PROXY_PORT:9001" \
        --variable "PROMETHEUS_HOST:$(hostname)" \
        suites/telemetry/telemetry.robot \
        suites/osconfig/clusterid.robot \
        suites/osconfig/systemd-resolved.robot
}
