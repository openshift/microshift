#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# Redefine network-related settings to use the dedicated network bridge
VM_BRIDGE_IP="$(get_vm_bridge_ip "${VM_MULTUS_NETWORK}")"
# shellcheck disable=SC2034  # used elsewhere
WEB_SERVER_URL="http://${VM_BRIDGE_IP}:${WEB_SERVER_PORT}"

start_image="rhel-9.6-microshift-brew-optionals-4.${MINOR_VERSION}-${LATEST_RELEASE_TYPE}"

# Skip the scenario if platform is ARM, as the igb driver is not supported.
check_platform() {
    if [[ "${UNAME_M}" =~ aarch64 ]] ; then
        record_junit "setup" "scenario_create_vms" "SKIPPED"
        exit 0
    fi
}

scenario_create_vms() {
    check_platform
    exit_if_commit_not_found "${start_image}"

    # Three nics - one for sriov, one for macvlan, another for ipvlan (they cannot enslave the same interface)
    LVM_SYSROOT_SIZE=20480 prepare_kickstart host1 kickstart.ks.template "${start_image}"
    launch_vm  --network "${VM_MULTUS_NETWORK},${VM_MULTUS_NETWORK},sriov" --vm_disksize 25
}

scenario_remove_vms() {
    check_platform
    exit_if_commit_not_found "${start_image}"

    remove_vm host1
}

scenario_run_tests() {
    check_platform
    exit_if_commit_not_found "${start_image}"

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
