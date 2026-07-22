#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# CIS hardening runs inside the Robot suite and takes ~20 minutes
# shellcheck disable=SC2034  # used elsewhere
TEST_EXECUTION_TIMEOUT=60m

# The RPM-based image used to create the VM for this test does not
# include MicroShift or greenboot, so tell the framework not to wait
# for greenboot to finish when creating the VM.
export SKIP_GREENBOOT=true

scenario_create_vms() {
    prepare_kickstart host1 kickstart-liveimg.ks.template ""
    launch_vm rhel102-installer

    configure_vm_firewall host1
    subscription_manager_register host1

    # Configure repositories for MicroShift and its dependencies.
    # MicroShift is NOT installed here — the Robot suite installs it
    # after CIS hardening so the scan reveals what MicroShift changes.
    local -r source_reponame=$(basename "${LOCAL_REPO}")
    local -r source_repo_url="${WEB_SERVER_URL}/rpm-repos/${source_reponame}"

    configure_rhocp_repo "${RHOCP_MINOR_Y}"       "${MAJOR_VERSION}" "${MINOR_VERSION}"
    configure_rhocp_repo "${RHOCP_MINOR_Y_BETA}"  "${MAJOR_VERSION}" "${MINOR_VERSION}"
    configure_rhocp_repo "${RHOCP_MINOR_Y1}"      "${PREVIOUS_MAJOR_VERSION}" "${PREVIOUS_MINOR_VERSION}"
    configure_rhocp_repo "${RHOCP_MINOR_Y1_BETA}" "${PREVIOUS_MAJOR_VERSION}" "${PREVIOUS_MINOR_VERSION}"
    configure_microshift_mirror "${PREVIOUS_RELEASE_REPO}"
    run_command_on_vm host1 "sudo subscription-manager release --set 10.2"
    configure_fast_datapath_repo

    # Add the source RPM repo for locally-built MicroShift packages.
    local -r tmp_file=$(mktemp)
    tee "${tmp_file}" >/dev/null <<EOF
[microshift-local]
name=MicroShift Local
baseurl=${source_repo_url}
enabled=1
gpgcheck=0
skip_if_unavailable=0
EOF
    copy_file_to_vm host1 "${tmp_file}" "${tmp_file}"
    run_command_on_vm host1 "sudo cp ${tmp_file} /etc/yum.repos.d/microshift-local.repo"
    rm -f "${tmp_file}"

}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    # shellcheck disable=SC2034  # used elsewhere
    TEST_RANDOMIZATION=none
    run_tests host1 \
        --variable "SCAP_DS_FILE:/usr/share/xml/scap/ssg/content/ssg-rhel10-ds.xml" \
        --variable "CIS_REQUIREMENTS_FILE:cis-requirements-el10.yml" \
        --variable "CIS_HARDEN_FILE:cis-harden-el10.yml" \
        suites/cis/
}
