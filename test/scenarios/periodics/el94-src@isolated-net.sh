#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# Redefine network-related settings to use the isolated network bridge
VM_BRIDGE_IP="$(get_vm_bridge_ip "${VM_ISOLATED_NETWORK}")"
# shellcheck disable=SC2034  # used elsewhere
WEB_SERVER_URL="http://${VM_BRIDGE_IP}:${WEB_SERVER_PORT}"

scenario_create_vms() {
    prepare_kickstart host1 kickstart-isolated.ks.template rhel-9.4-microshift-source-isolated
    # Use the isolated network when creating a VM
    launch_vm  --network_name "${VM_ISOLATED_NETWORK}"
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    run_tests host1 suites/network/isolated-network.robot
}
