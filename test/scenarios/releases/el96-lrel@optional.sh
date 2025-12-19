#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# Redefine network-related settings to use the dedicated network bridge
VM_BRIDGE_IP="$(get_vm_bridge_ip "${VM_MULTUS_NETWORK}")"
# shellcheck disable=SC2034  # used elsewhere
WEB_SERVER_URL="http://${VM_BRIDGE_IP}:${WEB_SERVER_PORT}"

start_image="rhel-9.6-microshift-brew-optionals-4.${MINOR_VERSION}-${LATEST_RELEASE_TYPE}"

scenario_create_vms() {
    exit_if_commit_not_found "${start_image}"

    # Skip sriov network on ARM because the igb driver is not supported.
    local networks="${VM_MULTUS_NETWORK},${VM_MULTUS_NETWORK},sriov"
    if [[ "${UNAME_M}" =~ aarch64 ]]; then
        networks="${VM_MULTUS_NETWORK},${VM_MULTUS_NETWORK}"
    fi
    # Three nics - one for sriov, one for macvlan, another for ipvlan (they cannot enslave the same interface)
    prepare_kickstart host1 kickstart.ks.template "${start_image}"
    launch_vm  --network "${networks}"
}

scenario_remove_vms() {
    exit_if_commit_not_found "${start_image}"

    remove_vm host1
}

scenario_run_tests() {
    exit_if_commit_not_found "${start_image}"

    local skip_args=""
    if [[ "${UNAME_M}" =~ aarch64 ]]; then
        skip_args="--skip sriov"
    fi
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
        ${skip_args} \
        suites/optional/
}
