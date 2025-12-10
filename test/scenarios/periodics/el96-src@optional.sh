#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# Redefine network-related settings to use the dedicated network bridge
VM_BRIDGE_IP="$(get_vm_bridge_ip "${VM_MULTUS_NETWORK}")"
# shellcheck disable=SC2034  # used elsewhere
WEB_SERVER_URL="http://${VM_BRIDGE_IP}:${WEB_SERVER_PORT}"

scenario_create_vms() {
    prepare_kickstart host1 kickstart.ks.template rhel-9.6-microshift-source-optionals
    # Two nics - one for macvlan, another for ipvlan (they cannot enslave the same interface)
    launch_vm  --network "${VM_MULTUS_NETWORK},${VM_MULTUS_NETWORK}"

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
        # Generic Device Plugin suite is excluded because getting serialsim for ostree would require:
        # - getting the version of the kernel of ostree image,
        # - installing kernel-devel of that version on the hypervisor,
        # - building serialsim
        # - packaging serialsim as an RPM
        # - including the RPM in the ostree blueprint
        # GDP suite is tested with bootc images instead.
        run_tests host1 \
        --variable "PROMETHEUS_HOST:$(hostname)" \
        --variable "LOKI_HOST:$(hostname)" \
        --exclude generic-device-plugin \
        suites/optional/
}
