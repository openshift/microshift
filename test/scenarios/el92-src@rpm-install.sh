#!/bin/bash

# shellcheck source=test/bin/get_rel_version_repo.sh
source "${SCRIPTDIR}/get_rel_version_repo.sh"

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
    launch_vm host1 "rhel-9.2"

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
    local -r target_version=$(current_version)
    local -r previous_minor_version=$(previous_minor_version)
    local -r previous_version_repo=$(get_rel_version_repo "${previous_minor_version}")
    local -r previous_version_repo_url=$(echo "${previous_version_repo}" | cut -d, -f2)

    # Enable the repositories with the dependencies using the most
    # recent GA OCP version. This value is used to build the repo
    # name, so should include the major and minor version number.
    local -r dependency_version="4.$("${ROOTDIR}/scripts/get-latest-rhocp-repo.sh")"

    run_tests host1 \
        --exitonfailure \
        --variable "SOURCE_REPO_URL:${source_repo_url}" \
        --variable "TARGET_VERSION:${target_version}" \
        --variable "PREVIOUS_MINOR_VERSION:${previous_minor_version}" \
        --variable "PREVIOUS_VERSION_REPO_URL:${previous_version_repo_url}" \
        --variable "DEPENDENCY_VERSION:${dependency_version}" \
        suites/rpm/install-and-upgrade-successful.robot
}
