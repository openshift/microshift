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
    prepare_kickstart host1 kickstart-liveimg.ks.template ""
    launch_vm rhel102-installer

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
    remove_vm host1
}

scenario_run_tests() {
    local -r source_reponame=$(basename "${LOCAL_REPO}")
    local -r source_repo_url="${WEB_SERVER_URL}/rpm-repos/${source_reponame}"
    local -r target_version=$(local_rpm_version)

    # Enable the rhocp and dependency repositories.
    #
    # Note that rhocp or beta dependencies repository may not yet exist
    # for the current version. Then, just use whatever we can get for
    # the previous minor version.
    configure_rhocp_repo "${RHOCP_MINOR_Y}"       4 "${MINOR_VERSION}"       "${arch}"
    configure_rhocp_repo "${RHOCP_MINOR_Y_BETA}"  4 "${MINOR_VERSION}"       "${arch}"
    configure_rhocp_repo "${RHOCP_MINOR_Y1}"      4 "${PREVIOUS_MINOR_VERSION}" "${arch}"
    configure_rhocp_repo "${RHOCP_MINOR_Y1_BETA}" 4 "${PREVIOUS_MINOR_VERSION}" "${arch}"
    configure_microshift_mirror "${PREVIOUS_RELEASE_REPO}"
    run_command_on_vm host1 "sudo subscription-manager release --set 10.2"
    configure_fast_datapath_repo

    run_tests host1 \
        --exitonfailure \
        --variable "SOURCE_REPO_URL:${source_repo_url}" \
        --variable "TARGET_VERSION:${target_version}" \
        --variable "PREVIOUS_MINOR_VERSION:${PREVIOUS_MINOR_VERSION}" \
        suites/rpm/install.robot \
        suites/rpm/remove.robot \
        suites/rpm/upgrade-successful.robot
}
