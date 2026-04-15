#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# Redefine network-related settings to use the dedicated IPv6 network bridge
# shellcheck disable=SC2034  # used elsewhere
VM_BRIDGE_IP="$(get_vm_bridge_ip "${VM_IPV6_NETWORK}")"
# shellcheck disable=SC2034  # used elsewhere
WEB_SERVER_URL="http://[${VM_BRIDGE_IP}]:${WEB_SERVER_PORT}"
# shellcheck disable=SC2034  # used elsewhere
MIRROR_REGISTRY_URL="${VM_BRIDGE_IP}:${MIRROR_REGISTRY_PORT}"

start_image="rhel98-brew-lrel-optional"
SCENARIO_VCPUS=4
SCENARIO_NETWORKS="${VM_IPV6_NETWORK}"

# Opt-in to dynamic VM scheduling by declaring requirements
dynamic_schedule_requirements() {
    echo "boot_image=${start_image}"
}

scenario_create_vms() {
    exit_if_commit_not_found "${start_image}"

    # Enable IPv6 single stack in kickstart
    prepare_kickstart host1 kickstart.ks.template "${start_image}" false true
    launch_vm rhel-9.8
}

scenario_remove_vms() {
    exit_if_commit_not_found "${start_image}"

    remove_vm host1
}

scenario_run_tests() {
    exit_if_commit_not_found "${start_image}"

    run_tests host1 suites/ipv6/singlestack.robot
}
