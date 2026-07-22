#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# CIS hardening runs inside the Robot suite and takes ~20 minutes
# shellcheck disable=SC2034  # used elsewhere
TEST_EXECUTION_TIMEOUT=60m

# The RPM-based image used to create the VM for this test does not
# include MicroShift or greenboot, so tell the framework not to wait
# for greenboot to finish when creating the VM.
export SKIP_GREENBOOT=true

# CIS tests must run in order: harden, scan baseline, install, scan post.
export TEST_RANDOMIZATION=none

scenario_create_vms() {
    prepare_kickstart host1 kickstart-liveimg.ks.template ""
    launch_vm rhel98-installer

    configure_vm_firewall host1
    subscription_manager_register host1

    # Configure repositories for MicroShift dependencies.
    # MicroShift itself is NOT installed here — the Robot suite configures
    # the local source repo and installs MicroShift after CIS hardening
    # so the scan reveals what MicroShift changes.
    configure_rhocp_repo "${RHOCP_MINOR_Y}"       "${MAJOR_VERSION}" "${MINOR_VERSION}"
    configure_rhocp_repo "${RHOCP_MINOR_Y_BETA}"  "${MAJOR_VERSION}" "${MINOR_VERSION}"
    configure_rhocp_repo "${RHOCP_MINOR_Y1}"      "${PREVIOUS_MAJOR_VERSION}" "${PREVIOUS_MINOR_VERSION}"
    configure_rhocp_repo "${RHOCP_MINOR_Y1_BETA}" "${PREVIOUS_MAJOR_VERSION}" "${PREVIOUS_MINOR_VERSION}"
    configure_microshift_mirror "${PREVIOUS_RELEASE_REPO}"
    run_command_on_vm host1 "sudo subscription-manager release --set 9.8"
    configure_fast_datapath_repo
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    local -r source_reponame=$(basename "${LOCAL_REPO}")
    local -r source_repo_url="${WEB_SERVER_URL}/rpm-repos/${source_reponame}"
    local -r target_version=$(local_rpm_version)

    run_tests host1 \
        --variable "SOURCE_REPO_URL:${source_repo_url}" \
        --variable "TARGET_VERSION:${target_version}" \
        suites/cis/
}
