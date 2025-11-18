#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# Redefine network-related settings to use the dedicated IPv6 network bridge
# shellcheck disable=SC2034  # used elsewhere
VM_BRIDGE_IP="$(get_vm_bridge_ip "${VM_IPV6_NETWORK}")"
# shellcheck disable=SC2034  # used elsewhere
WEB_SERVER_URL="http://[${VM_BRIDGE_IP}]:${WEB_SERVER_PORT}"

start_image="rhel96-bootc-brew-${LATEST_RELEASE_TYPE}-with-optional"

scenario_create_vms() {
    exit_if_image_not_found "${start_image}"

    # Using `hostname` here instead of a raw ip because skopeo only allows either
    # ipv4 or fqdn's, but not ipv6. Since the registry is hosted on the ipv6
    # network gateway in the host, we need to use a combination of the hostname
    # plus /etc/hosts resolution (which is taken care of by kickstart).
    # Save the original value and temporarily override for prepare_kickstart
    local original_mirror_registry_url="${MIRROR_REGISTRY_URL}"
    MIRROR_REGISTRY_URL="$(hostname):${MIRROR_REGISTRY_PORT}/microshift"
    # Enable IPv6 single stack in kickstart
    prepare_kickstart host1 kickstart-bootc.ks.template "${start_image}" false true
    MIRROR_REGISTRY_URL="${original_mirror_registry_url}"
    launch_vm --boot_blueprint rhel96-bootc --network "${VM_IPV6_NETWORK}"
}

scenario_remove_vms() {
    exit_if_image_not_found "${start_image}"

    remove_vm host1
}

scenario_run_tests() {
    exit_if_image_not_found "${start_image}"

    run_tests host1 suites/ipv6/singlestack.robot
}
