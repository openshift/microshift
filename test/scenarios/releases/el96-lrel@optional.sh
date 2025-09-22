#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# Redefine network-related settings to use the dedicated network bridge
VM_BRIDGE_IP="$(get_vm_bridge_ip "${VM_MULTUS_NETWORK}")"
# shellcheck disable=SC2034  # used elsewhere
WEB_SERVER_URL="http://${VM_BRIDGE_IP}:${WEB_SERVER_PORT}"

scenario_create_vms() {
    prepare_kickstart host1 kickstart.ks.template "rhel-9.6-microshift-brew-optionals-4.${MINOR_VERSION}-${LATEST_RELEASE_TYPE}"
    # Two nics - one for macvlan, another for ipvlan (they cannot enslave the same interface)
    launch_vm  --network "${VM_MULTUS_NETWORK},${VM_MULTUS_NETWORK}"
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
        --variable "PROMETHEUS_PORT:9092" \
        --variable "LOKI_HOST:$(hostname)" \
        --variable "LOKI_PORT:3200" \
        --variable "PROM_EXPORTER_PORT:8889" \
        --exclude generic-device-plugin \
        suites/optional/
}
