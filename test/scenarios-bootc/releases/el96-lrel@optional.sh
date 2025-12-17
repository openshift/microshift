#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# Redefine network-related settings to use the dedicated network bridge
VM_BRIDGE_IP="$(get_vm_bridge_ip "${VM_MULTUS_NETWORK}")"
# shellcheck disable=SC2034  # used elsewhere
WEB_SERVER_URL="http://${VM_BRIDGE_IP}:${WEB_SERVER_PORT}"

start_image="rhel96-bootc-brew-${LATEST_RELEASE_TYPE}-with-optional"

# Skip the scenario if platform is ARM, as the igb driver is not supported.
check_platform() {
    if [[ "${UNAME_M}" =~ aarch64 ]] ; then
        record_junit "setup" "scenario_create_vms" "SKIPPED"
        exit 0
    fi
}

scenario_create_vms() {
    check_platform
    exit_if_image_not_found "${start_image}"

    prepare_kickstart host1 kickstart-bootc.ks.template "${start_image}"
    # Three nics - one for sriov, one for macvlan, another for ipvlan (they cannot enslave the same interface)
    launch_vm --boot_blueprint rhel96-bootc --network "${VM_MULTUS_NETWORK},${VM_MULTUS_NETWORK},sriov"
}

scenario_remove_vms() {
    check_platform
    exit_if_image_not_found "${start_image}"

    remove_vm host1
}

scenario_run_tests() {
    check_platform
    exit_if_image_not_found "${start_image}"

    run_tests host1 \
        --variable "PROMETHEUS_HOST:$(hostname)" \
        --variable "LOKI_HOST:$(hostname)" \
        suites/optional/
}
