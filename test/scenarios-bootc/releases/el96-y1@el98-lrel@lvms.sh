#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# NOTE: Unlike most suites, these tests rely on being run IN ORDER to
# ensure MicroShift is upgraded before running validation tests
export TEST_RANDOMIZATION=none

start_image="rhel96-bootc-brew-y1-with-optional"
dest_image="rhel98-bootc-brew-lrel-optional"

scenario_create_vms() {
    exit_if_image_not_found "${start_image}"
    exit_if_image_not_found "${dest_image}"

    prepare_kickstart host1 kickstart-bootc.ks.template "${start_image}"
    launch_vm --boot_blueprint rhel96-bootc --vm_vcpus 4
}

scenario_remove_vms() {
    exit_if_image_not_found "${start_image}"
    exit_if_image_not_found "${dest_image}"

    remove_vm host1
}

scenario_run_tests() {
    exit_if_image_not_found "${start_image}"
    exit_if_image_not_found "${dest_image}"

    # Wait for MicroShift to be ready
    wait_for_microshift_to_be_ready host1

    # Setup oc client and kubeconfig for ginkgo tests
    setup_oc_and_kubeconfig host1

    # Pre-upgrade: Create LVMS workloads and validate LVMS is working
    echo "INFO: Creating LVMS workloads before upgrade..."
    run_command_on_vm host1 'bash -s' < "${TESTDIR}/../scripts/lvms-helpers/createWorkloads.sh"

    echo "INFO: Checking LVMS resources before upgrade..."
    run_command_on_vm host1 'bash -s' < "${TESTDIR}/../scripts/lvms-helpers/checkLvmsResources.sh"

    # Perform upgrade and validate
    run_tests host1 \
        --variable "TARGET_REF:${dest_image}" \
        --variable "BOOTC_REGISTRY:${MIRROR_REGISTRY_URL}" \
        suites/upgrade/upgrade-successful.robot

    # Post-upgrade: Validate LVMS workloads survived the upgrade
    echo "INFO: Checking LVMS workloads survived upgrade..."
    run_command_on_vm host1 'bash -s' < "${TESTDIR}/../scripts/lvms-helpers/checkWorkloadExists.sh"

    echo "INFO: Checking LVMS resources after upgrade..."
    run_command_on_vm host1 'bash -s' < "${TESTDIR}/../scripts/lvms-helpers/checkLvmsResources.sh"

    # Cleanup LVMS workloads
    echo "INFO: Cleaning up LVMS workloads..."
    run_command_on_vm host1 'bash -s' < "${TESTDIR}/../scripts/lvms-helpers/cleanupWorkload.sh"

    # Run standard1 suite for basic validation after upgrade
    run_tests host1 \
        --variable "EXPECTED_OS_VERSION:9.8" \
        suites/standard1/
}
