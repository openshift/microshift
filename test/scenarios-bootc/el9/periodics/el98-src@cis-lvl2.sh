#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# CIS hardening runs inside the Robot suite and takes ~20 minutes
# shellcheck disable=SC2034  # used elsewhere
TEST_EXECUTION_TIMEOUT=60m

# The RPM-based image used to create the VM for this test does not
# include MicroShift or greenboot, so tell the framework not to wait
# for greenboot to finish when creating the VM.
export SKIP_GREENBOOT=true

configure_microshift_mirror() {
    local -r repo=$1

    if [[ -z "${repo}" ]] ; then
        return
    fi

    if [[ ! "${repo}" =~ ^http ]]; then
        return
    fi

    local -r tmp_file=$(mktemp)
    tee "${tmp_file}" >/dev/null <<EOF
[microshift-mirror-rpms]
name=MicroShift Mirror
baseurl=${repo}
enabled=1
gpgcheck=0
skip_if_unavailable=0
EOF
    copy_file_to_vm host1 "${tmp_file}" "${tmp_file}"
    run_command_on_vm host1 "sudo cp ${tmp_file} /etc/yum.repos.d/microshift-mirror-rpms.repo"
    rm -f "${tmp_file}"
}

configure_rhocp_repo() {
    local -r rhocp=$1
    local -r major=$2
    local -r minor=$3

    if [[ -z "${rhocp}" ]] ; then
        return
    fi

    if [[ "${rhocp}" =~ ^[0-9]{1,2}$ ]]; then
        run_command_on_vm host1 "sudo subscription-manager repos --enable rhocp-${major}.${rhocp}-for-rhel-9-\$(uname -m)-rpms"
    elif [[ "${rhocp}" =~ ^http ]]; then
        local -r ocp_repo_name="rhocp-${major}.${minor}-for-rhel-9-mirrorbeta-rpms"
        local -r tmp_file=$(mktemp)

        tee "${tmp_file}" >/dev/null <<EOF
[${ocp_repo_name}]
name=Beta rhocp RPMs for RHEL 9
baseurl=${rhocp}
enabled=1
gpgcheck=0
skip_if_unavailable=0
EOF
        copy_file_to_vm host1 "${tmp_file}" "${tmp_file}"
        run_command_on_vm host1 "sudo cp ${tmp_file} /etc/yum.repos.d/${ocp_repo_name}.repo"
        rm -f "${tmp_file}"
    fi
}

scenario_create_vms() {
    prepare_kickstart host1 kickstart-liveimg.ks.template ""
    launch_vm rhel98-installer

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
    run_command_on_vm host1 "sudo subscription-manager release --set 9.8"
    run_command_on_vm host1 "sudo subscription-manager repos --enable fast-datapath-for-rhel-9-\$(uname -m)-rpms"

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
        suites/cis/
}
