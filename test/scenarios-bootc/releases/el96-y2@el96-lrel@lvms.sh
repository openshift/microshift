#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# NOTE: Unlike most suites, these tests rely on being run IN ORDER to
# ensure MicroShift is upgraded before running validation tests
export TEST_RANDOMIZATION=none

start_image="rhel96-bootc-brew-y2-with-optional"
dest_image="rhel96-bootc-brew-${LATEST_RELEASE_TYPE}-with-optional"

scenario_create_vms() {
    if ! does_image_exist "${start_image}"; then
        echo "Image '${start_image}' not found - skipping test"
        return 0
    fi
    if ! does_image_exist "${dest_image}"; then
        echo "Image '${dest_image}' not found - skipping test"
        return 0
    fi
    prepare_kickstart host1 kickstart-bootc.ks.template "${start_image}"
    launch_vm --boot_blueprint rhel96-bootc
}

scenario_remove_vms() {
    if ! does_image_exist "${start_image}"; then
        echo "Image '${start_image}' not found - skipping test"
        return 0
    fi
    if ! does_image_exist "${dest_image}"; then
        echo "Image '${dest_image}' not found - skipping test"
        return 0
    fi
    remove_vm host1
}

scenario_run_tests() {
    if ! does_image_exist "${start_image}"; then
        echo "Image '${start_image}' not found - skipping test"
        return 0
    fi

    if ! does_image_exist "${dest_image}"; then
        echo "Image '${dest_image}' not found - skipping test"
        return 0
    fi

    # Setup oc client and kubeconfig for gingko tests
    setup_oc_and_kubeconfig_tests host1

    # Wait for MicroShift to be ready
    wait_for_microshift_to_be_ready host1

    # Pre-upgrade: Create LVMS workloads and validate LVMS is working
    echo "INFO: Creating LVMS workloads before upgrade..."
    run_command_on_vm host1 'bash -s' < "${TESTDIR}/../scripts/lvms-helpers/createWorkloads.sh"

    echo "INFO: Checking LVMS resources before upgrade..."
    run_command_on_vm host1 'bash -s' < "${TESTDIR}/../scripts/lvms-helpers/checkLvmsResources.sh"

    # Perform upgrade and validate
    run_tests host1 \
        --variable "TARGET_REF:${dest_image}" \
        --variable "BOOTC_REGISTRY:${MIRROR_REGISTRY_URL}" \
        --variable "EXPECTED_OS_VERSION:9.6" \
        suites/upgrade/upgrade-successful.robot

    # Post-upgrade: Validate LVMS workloads survived the upgrade
    echo "INFO: Checking LVMS workloads survived upgrade..."
    run_command_on_vm host1 'bash -s' < "${TESTDIR}/../scripts/lvms-helpers/checkWorkloadExists.sh"

    echo "INFO: Checking LVMS resources after upgrade..."
    run_command_on_vm host1 'bash -s' < "${TESTDIR}/../scripts/lvms-helpers/checkLvmsResources.sh"

    # Run ginkgo tests to validate functionality
    run_gingko_tests host1 "~Disruptive"

    # Cleanup LVMS workloads
    echo "INFO: Cleaning up LVMS workloads..."
    run_command_on_vm host1 'bash -s' < "${TESTDIR}/../scripts/lvms-helpers/cleanupWorkload.sh"
}
