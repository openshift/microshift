#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# shellcheck disable=SC2034  # used elsewhere
# Increase greenboot timeout for optional packages (more services to start)
GREENBOOT_TIMEOUT=1200

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
    LVM_SYSROOT_SIZE=20480 prepare_kickstart host1 kickstart-bootc.ks.template rhel96-bootc-source-optionals
    # Three nics - one for sriov, one for macvlan, another for ipvlan (they cannot enslave the same interface)
    launch_vm --boot_blueprint rhel96-bootc --network "${networks}" --vm_disksize 25 --vm_vcpus 4
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    local skip_args=""
    if [[ "${UNAME_M}" =~ aarch64 ]]; then
        skip_args="--skip sriov"
    fi
    # shellcheck disable=SC2086
    run_tests host1 \
        --variable "PROMETHEUS_HOST:$(hostname)" \
        --variable "LOKI_HOST:$(hostname)" \
        ${skip_args} \
        suites/optional/
}
