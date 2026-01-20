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

configure_rhocp_repo() {
    local -r rhocp=$1
    local -r version=$2

    # The repository may be empty if the beta mirror is not up yet
    if [[ -z "${rhocp}" ]] ; then
        return
    fi

    if [[ "${rhocp}" =~ ^[0-9]{2} ]]; then
        run_command_on_vm host1 "sudo subscription-manager repos --enable rhocp-4.${rhocp}-for-rhel-9-\$(uname -m)-rpms"
    elif [[ "${rhocp}" =~ ^http ]]; then
        local -r ocp_repo_name="rhocp-4.${version}-for-rhel-9-mirrorbeta-rpms"
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
    remove_vm host1
}

scenario_run_tests() {
    local -r reponame=$(basename "${BREW_REPO}")
    local -r repo_url="${WEB_SERVER_URL}/rpm-repos/${reponame}"

    # Enable the rhocp and dependency repositories.
    #
    # Note that rhocp or beta dependencies repository may not yet exist
    # for the current version. Then, just use whatever we can get for
    # the previous minor version.
    configure_rhocp_repo "${RHOCP_MINOR_Y}"       "${MINOR_VERSION}"
    configure_rhocp_repo "${RHOCP_MINOR_Y_BETA}"  "${MINOR_VERSION}"
    run_command_on_vm host1 "sudo subscription-manager repos --enable fast-datapath-for-rhel-9-\$(uname -m)-rpms"

    run_tests host1 \
        --exitonfailure \
        --variable "SOURCE_REPO_URL:${repo_url}" \
        --variable "TARGET_VERSION:${BREW_LATEST_RELEASE_VERSION}" \
        --variable "EXPECTED_OS_VERSION:9.6" \
        suites/rpm/install.robot \
        suites/standard1/ suites/selinux/validate-selinux-policy.robot
}
