#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# Redefine network-related settings to use the dedicated network bridge
VM_BRIDGE_IP="$(get_vm_bridge_ip "${VM_MULTUS_NETWORK}")"
# shellcheck disable=SC2034  # used elsewhere
WEB_SERVER_URL="http://${VM_BRIDGE_IP}:${WEB_SERVER_PORT}"

scenario_create_vms() {
    prepare_kickstart host1 kickstart-bootc.ks.template rhel94-bootc-source-optionals
    # Using centos9 is necessary for getting the latest anaconda.
    # It is a temporary workaround until rhel-9.4.iso build is available.
    # Two nics - one for macvlan, another for ipvlan (they cannot enslave the same interface)
    launch_vm host1 centos9 "${VM_MULTUS_NETWORK}" "" "" "" "2" "" "1"

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
    run_tests host1 suites/optional/
}
