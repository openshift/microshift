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
    launch_vm host1 "rhel-9.4"

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
    local -r current_minor_version=$(current_minor_version)
    local -r previous_minor_version=$(previous_minor_version)

    # Enable the repositories with the dependencies.

    # RHOCP for current version might not exist yet. Then, just use whatever we can get for previous minor.
    local -r rhocp_current=$("${ROOTDIR}/scripts/get-latest-rhocp-repo.sh" "${current_minor_version}" || true)
    local -r rhocp_previous=$("${ROOTDIR}/scripts/get-latest-rhocp-repo.sh" "${previous_minor_version}")
    for rhocp in "${rhocp_current}" "${rhocp_previous}"; do
        if [[ ! "${rhocp}" ]]; then
            # rhocp_current might be empty if the beta mirror is not up yet
            continue
        fi

        if [[ "${rhocp}" =~ ^[0-9]{2} ]]; then
            run_command_on_vm host1 "sudo subscription-manager repos --enable rhocp-4.${rhocp}-for-rhel-9-\$(uname -m)-rpms"
        elif [[ "${rhocp}" =~ ^http ]]; then
            url=$(echo "${rhocp}" | cut -d, -f1)
            ver=$(echo "${rhocp}" | cut -d, -f2)
            OCP_REPO_NAME="rhocp-4.${ver}-for-rhel-9-mirrorbeta-$(uname -i)-rpms"
            tmp_file=$(mktemp)
            tee "${tmp_file}" >/dev/null <<EOF
[${OCP_REPO_NAME}]
name=Beta rhocp-4.${ver} RPMs for RHEL 9
baseurl=${url}
enabled=1
gpgcheck=0
skip_if_unavailable=0
EOF
            copy_file_to_vm host1 "${tmp_file}" "${tmp_file}"
            run_command_on_vm host1 "sudo cp ${tmp_file} /etc/yum.repos.d/${OCP_REPO_NAME}.repo"
            rm "${tmp_file}"
        fi
    done

    run_command_on_vm host1 "sudo subscription-manager repos --enable fast-datapath-for-rhel-9-\$(uname -m)-rpms"

    run_tests host1 \
        --exitonfailure \
        --variable "SOURCE_REPO_URL:${source_repo_url}" \
        --variable "TARGET_VERSION:${target_version}" \
        --variable "PREVIOUS_MINOR_VERSION:${previous_minor_version}" \
        suites/rpm/install-and-upgrade-successful.robot
}
