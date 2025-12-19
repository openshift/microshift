#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# Redefine network-related settings to use the dedicated network bridge
VM_BRIDGE_IP="$(get_vm_bridge_ip "${VM_MULTUS_NETWORK}")"
# shellcheck disable=SC2034  # used elsewhere
WEB_SERVER_URL="http://${VM_BRIDGE_IP}:${WEB_SERVER_PORT}"

scenario_create_vms() {
    # Skip sriov network on ARM because the igb driver is not supported.
    local networks="${VM_MULTUS_NETWORK},${VM_MULTUS_NETWORK},sriov"
    if [[ "${UNAME_M}" =~ aarch64 ]]; then
        networks="${VM_MULTUS_NETWORK},${VM_MULTUS_NETWORK}"
    fi
    prepare_kickstart host1 kickstart-bootc.ks.template rhel96-bootc-source-optionals
    # Three nics - one for sriov, one for macvlan, another for ipvlan (they cannot enslave the same interface)
    launch_vm --boot_blueprint rhel96-bootc --network "${networks}"

    # Open the firewall ports. Other scenarios get this behavior by
    # embedding settings in the blueprint, but there is no blueprint
    # for this scenario. We need do this step before running the RF
    # suite so that suite can assume it can reach all of the same
    # ports as for any other test.
    configure_vm_firewall host1
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    local skip_args=""
    if [[ "${UNAME_M}" =~ aarch64 ]]; then
        skip_args="--skip sriov"
    fi
    run_tests host1 \
        --variable "PROMETHEUS_HOST:$(hostname)" \
        --variable "LOKI_HOST:$(hostname)" \
        "${skip_args}" \
        suites/optional/
}
