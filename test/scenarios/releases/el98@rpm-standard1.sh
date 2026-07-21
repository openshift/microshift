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
    exit_if_brew_rpms_not_found

    prepare_kickstart host1 kickstart-liveimg.ks.template ""
    launch_vm rhel-9.8
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
    exit_if_brew_rpms_not_found

    remove_vm host1
}

scenario_run_tests() {
    exit_if_brew_rpms_not_found

    local -r reponame=$(basename "${BREW_REPO}")
    local -r repo_url="${WEB_SERVER_URL}/rpm-repos/${reponame}"

    # Enable the rhocp and dependency repositories.
    #
    # Note that rhocp or beta dependencies repository may not yet exist
    # for the current version. Then, just use whatever we can get for
    # the previous minor version.
    configure_rhocp_repo "${RHOCP_MINOR_Y}"       "${MINOR_VERSION}"
    configure_rhocp_repo "${RHOCP_MINOR_Y_BETA}"  "${MINOR_VERSION}"
    run_command_on_vm host1 "sudo subscription-manager release --set 9.8"
    configure_fast_datapath_repo

    run_tests host1 \
        --exitonfailure \
        --variable "SOURCE_REPO_URL:${repo_url}" \
        --variable "TARGET_VERSION:${BREW_LREL_RELEASE_VERSION}" \
        --variable "EXPECTED_OS_VERSION:9.8" \
        suites/rpm/install.robot \
        suites/standard1/ \
        suites/rpm/remove.robot
}
