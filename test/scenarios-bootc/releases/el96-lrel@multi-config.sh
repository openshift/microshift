#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# Multi-config scenario: Combines multiple configurations to validate no conflicts
# - Low-latency (tuned)
# - TLSv1.3
# - LVMS (default)
# - IPv6 (network configuration)

export SKIP_GREENBOOT=true
export TEST_RANDOMIZATION=none

# Redefine network-related settings to use the dedicated IPv6 network bridge
# shellcheck disable=SC2034  # used elsewhere
VM_BRIDGE_IP="$(get_vm_bridge_ip "${VM_IPV6_NETWORK}")"
# shellcheck disable=SC2034  # used elsewhere
WEB_SERVER_URL="http://[${VM_BRIDGE_IP}]:${WEB_SERVER_PORT}"

start_image="rhel96-bootc-brew-${LATEST_RELEASE_TYPE}-with-optional-tuned"

scenario_create_vms() {
    if ! does_image_exist "${start_image}"; then
        echo "Image '${start_image}' not found - skipping test"
        return 0
    fi

    # Temporarily override MIRROR_REGISTRY_URL for kickstart preparation
    # The kickstart template needs a hostname-based URL, not an IPv6 address
    local original_mirror_url="${MIRROR_REGISTRY_URL}"
    # shellcheck disable=SC2034  # used elsewhere
    MIRROR_REGISTRY_URL="$(hostname):${MIRROR_REGISTRY_PORT}/microshift"

    # Enable IPv6 single stack in kickstart
    prepare_kickstart host1 kickstart-bootc.ks.template "${start_image}" false true

    # Restore original MIRROR_REGISTRY_URL for runtime use
    # shellcheck disable=SC2034  # used elsewhere
    MIRROR_REGISTRY_URL="${original_mirror_url}"

    launch_vm --boot_blueprint rhel96-bootc --network "${VM_IPV6_NETWORK}" --vm_vcpus 6
}

scenario_remove_vms() {
    if ! does_image_exist "${start_image}"; then
        echo "Image '${start_image}' not found - skipping test"
        return 0
    fi

    remove_vm host1
}

scenario_run_tests() {
    if ! does_image_exist "${start_image}"; then
        echo "Image '${start_image}' not found - skipping test"
        return 0
    fi

    # Wait for microshift-tuned to reboot the node
    local -r start_time=$(date +%s)
    while true; do
        boot_num=$(run_command_on_vm host1 "sudo journalctl --list-boots --quiet | wc -l" || true)
        boot_num="${boot_num%$'\r'*}"
        if [[ "${boot_num}" -ge 2 ]]; then
            break
        fi
        if [ $(( $(date +%s) - start_time )) -gt 60 ]; then
            echo "Timed out waiting for VM having 2 boots"
            exit 1
        fi
        sleep 5
    done

    # Apply TLSv1.3 configuration via drop-in config
    echo "INFO: Configuring TLSv1.3..."
    run_command_on_vm host1 "sudo mkdir -p /etc/microshift/config.d"
    run_command_on_vm host1 "sudo tee /etc/microshift/config.d/10-tls.yaml > /dev/null << 'EOF'
apiServer:
  tls:
    minVersion: VersionTLS13
EOF"

    # Restart MicroShift to apply TLS configuration
    run_command_on_vm host1 "sudo systemctl restart microshift"

    # Setup oc client and kubeconfig for gingko tests
    setup_oc_and_kubeconfig host1

    # Create LVMS workloads
    echo "INFO: Creating LVMS workloads to validate storage..."
    run_command_on_vm host1 'bash -s' < "${TESTDIR}/../scripts/lvms-helpers/createWorkloads.sh"

    echo "INFO: Checking LVMS resources..."
    run_command_on_vm host1 'bash -s' < "${TESTDIR}/../scripts/lvms-helpers/checkLvmsResources.sh"

    # Run all tests in a single run_tests call to validate all configurations work together
    echo "INFO: Running validation tests for multi-config scenario..."
    run_tests host1 \
        --variable "EXPECTED_OS_VERSION:9.6" \
        --exclude tls-configuration \
        suites/ipv6/singlestack.robot \

    # Validate LVMS still works after all tests
    echo "INFO: Validating LVMS workloads after tests..."
    run_command_on_vm host1 'bash -s' < "${TESTDIR}/../scripts/lvms-helpers/checkWorkloadExists.sh"

    # Cleanup LVMS workloads
    echo "INFO: Cleaning up LVMS workloads..."
    run_command_on_vm host1 'bash -s' < "${TESTDIR}/../scripts/lvms-helpers/cleanupWorkload.sh"

    echo "SUCCESS: Multi-config scenario validation completed - no conflicts detected"
}
