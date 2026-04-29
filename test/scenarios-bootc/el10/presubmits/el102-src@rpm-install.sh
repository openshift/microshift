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

configure_microshift_mirror() {
    local -r repo=$1

    # `repo` might be empty if we install microshift from rhocp
    if [[ -z "${repo}" ]] ; then
        return
    fi

    # `repo` might be an enabled repo from a released version instead
    # of a mirror.
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

# On RHEL 10, rhocp and fast-datapath repos are not available via
# subscription-manager. Create repo files pointing to the RHEL 9 CDN
# using entitlement certificates as a workaround.
configure_cdn_repo() {
    local -r repo_id=$1
    local -r repo_name=$2
    local -r baseurl=$3

    local -r cert=$(run_command_on_vm host1 "ls /etc/pki/entitlement/[0-9]*.pem | grep -v '\-key.pem' | head -n1")
    local -r key=$(run_command_on_vm host1 "ls /etc/pki/entitlement/[0-9]*-key.pem | head -n1")
    local -r tmp_file=$(mktemp)

    tee "${tmp_file}" >/dev/null <<EOF
[${repo_id}]
name=${repo_name}
baseurl=${baseurl}
enabled=1
gpgcheck=1
gpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-redhat-release
sslverify=1
sslcacert=/etc/rhsm/ca/redhat-uep.pem
sslclientcert=${cert}
sslclientkey=${key}
EOF
    copy_file_to_vm host1 "${tmp_file}" "${tmp_file}"
    run_command_on_vm host1 "sudo cp ${tmp_file} /etc/yum.repos.d/${repo_id}.repo"
    rm -f "${tmp_file}"
}

configure_rhocp_repo() {
    local -r rhocp=$1
    local -r major=$2
    local -r minor=$3
    local -r arch=$4

    # The repository may be empty if the beta mirror is not up yet
    if [[ -z "${rhocp}" ]] ; then
        return
    fi

    if [[ "${rhocp}" =~ ^[0-9]{1,2}$ ]]; then
        configure_cdn_repo \
            "rhocp-${major}.${rhocp}" \
            "Red Hat OpenShift ${major}.${rhocp} for RHEL 9" \
            "https://cdn.redhat.com/content/dist/layered/rhel9/${arch}/rhocp/${major}.${rhocp}/os"
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
    local -r arch=$(uname -m)

    # Enable the rhocp and dependency repositories.
    #
    # Note that rhocp or beta dependencies repository may not yet exist
    # for the current version. Then, just use whatever we can get for
    # the previous minor version.
    configure_rhocp_repo "${RHOCP_MINOR_Y}"       "${MAJOR_VERSION}" "${MINOR_VERSION}"
    configure_rhocp_repo "${RHOCP_MINOR_Y_BETA}"  "${MAJOR_VERSION}" "${MINOR_VERSION}"
    configure_rhocp_repo "${RHOCP_MINOR_Y1}"      "${PREVIOUS_MAJOR_VERSION}" "${PREVIOUS_MINOR_VERSION}"
    configure_rhocp_repo "${RHOCP_MINOR_Y1_BETA}" "${PREVIOUS_MAJOR_VERSION}" "${PREVIOUS_MINOR_VERSION}"
    configure_microshift_mirror "${PREVIOUS_RELEASE_REPO}"
    configure_cdn_repo \
        "fast-datapath" \
        "Red Hat Fast Datapath for RHEL 9" \
        "https://cdn.redhat.com/content/dist/layered/rhel9/${arch}/fast-datapath/os"

    run_tests host1 \
        --exitonfailure \
        --variable "SOURCE_REPO_URL:${source_repo_url}" \
        --variable "TARGET_VERSION:${target_version}" \
        --variable "PREVIOUS_MINOR_VERSION:${PREVIOUS_MINOR_VERSION}" \
        suites/rpm/install.robot \
        suites/rpm/remove.robot \
        suites/rpm/upgrade-successful.robot
}
