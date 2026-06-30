#!/bin/bash
# Common helper functions for RPM-install scenarios.
# Sourced by individual scenario scripts after scenario.sh sets up the environment.

export SKIP_GREENBOOT=true
export TEST_RANDOMIZATION=none

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

wait_for_microshift_endpoint() {
    local -r endpoint=$1
    local -r kubeconfig="/var/lib/microshift/resources/kubeadmin/kubeconfig"
    local attempt
    for attempt in $(seq 30); do
        if run_command_on_vm host1 "sudo /usr/bin/oc get --kubeconfig ${kubeconfig} --raw='${endpoint}' 2>/dev/null" | grep -q "ok"; then
            return 0
        fi
        echo "Waiting for MicroShift ${endpoint} (attempt ${attempt}/30)"
        sleep 10
    done
    echo "ERROR: MicroShift ${endpoint} did not become ready"
    return 1
}

install_microshift() {
    local -r repo_url=$1
    local -r target_version=$2

    local -r tmp_file=$(mktemp)
    tee "${tmp_file}" >/dev/null <<EOF
[microshift-local]
name=MicroShift Local Repository
baseurl=${repo_url}
enabled=1
gpgcheck=0
EOF
    copy_file_to_vm host1 "${tmp_file}" "${tmp_file}"
    run_command_on_vm host1 "sudo cp ${tmp_file} /etc/yum.repos.d/microshift-local.repo"
    rm -f "${tmp_file}"

    run_command_on_vm host1 "sudo dnf install -q -R 2 -y --allowerasing 'microshift-${target_version}'"
    # Keep the local repo for scenarios that install additional RPMs (e.g. microshift-low-latency)
    # Wait for NetworkManager to reconnect after the RPM %post scriptlet restarts it
    run_command_on_vm host1 "sudo nmcli -w 30 networking connectivity check || true"
    run_command_on_vm host1 "sudo systemctl start microshift.service"

    wait_for_microshift_endpoint /readyz
    wait_for_microshift_endpoint /livez
    echo "MicroShift is ready"
}

configure_rhocp_repo() {
    local -r rhocp=$1
    local -r major=$2
    local -r minor=$3
    local -r arch=${4:-$(uname -m)}

    if [[ -z "${rhocp}" ]] ; then
        return
    fi

    if [[ "${rhocp}" =~ ^[0-9]{1,2}$ ]]; then
        if [[ "${RPM_RHEL_MAJOR}" -ge 10 ]]; then
            configure_cdn_repo \
                "rhocp-${major}.${rhocp}" \
                "Red Hat OpenShift ${major}.${rhocp} for RHEL 9" \
                "https://cdn.redhat.com/content/dist/layered/rhel9/${arch}/rhocp/${major}.${rhocp}/os"
        else
            run_command_on_vm host1 "sudo subscription-manager repos --enable rhocp-${major}.${rhocp}-for-rhel-9-\$(uname -m)-rpms"
        fi
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
