#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# The RPM-based image used to create the VM for this test does not
# include MicroShift or greenboot, so tell the framework not to wait
# for greenboot to finish when creating the VM.
export SKIP_GREENBOOT=true

# NOTE: Unlike most suites, these tests rely on being run IN ORDER to
# ensure the host is in a good state at the start of each test. We
# could have separated them and run them as separate scenarios, but
# did not want to spend the resources on a new VM.
export TEST_RANDOMIZATION=none

scenario_create_vms() {
    exit_if_zprev_not_exist

    prepare_kickstart host1 kickstart-liveimg.ks.template ""
    launch_vm --boot_blueprint rhel-9.6
    # Open the firewall ports. Other scenarios get this behavior by
    # embedding settings in the blueprint, but there is no blueprint
    # for this scenario. We need do this step before running the RF
    # suite so that suite can assume it can reach all of the same
    # ports as for any other test.
    configure_vm_firewall host1

    # Register the host with subscription manager so we can install
    # dependencies.
    subscription_manager_register host1
}

scenario_remove_vms() {
    exit_if_zprev_not_exist

    remove_vm host1
}

scenario_run_tests() {
    exit_if_zprev_not_exist

    # Enable the rhocp repo for the current minor version to install previous Z-stream version
    run_command_on_vm host1 "sudo subscription-manager repos --enable fast-datapath-for-rhel-9-\$(uname -m)-rpms"
    run_command_on_vm host1 "sudo subscription-manager repos --enable rhocp-4.${MINOR_VERSION}-for-rhel-9-\$(uname -m)-rpms"

    local -r reponame=$(basename "${BREW_REPO}")
    local -r repo_url="${WEB_SERVER_URL}/rpm-repos/${reponame}"

    run_tests host1 \
        --exitonfailure \
        --variable "SOURCE_REPO_URL:${repo_url}" \
        --variable "TARGET_VERSION:${BREW_LREL_RELEASE_VERSION}" \
        --variable "PREVIOUS_MINOR_VERSION:${MINOR_VERSION}" \
        suites/rpm/upgrade-successful.robot
}
