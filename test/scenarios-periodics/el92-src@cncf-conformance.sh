#!/bin/bash

# Sourced from cleanup_scenario.sh and uses functions defined there.

KUBECONFIG="${SCENARIO_INFO_DIR}/${SCENARIO}/kubeconfig"
# Timeout in seconds
TIMEOUT_TEST=7400
TIMEOUT_RESULTS=600

prepare_hosts() {
    local -r primary_host_ip=$(cat "${SCENARIO_INFO_DIR}/${SCENARIO}/vms/host1/public_ip")
    local -r secondary_host_ip=$(cat "${SCENARIO_INFO_DIR}/${SCENARIO}/vms/host2/public_ip")
    local -r primary_host_ssh_port=$(cat "${SCENARIO_INFO_DIR}/${SCENARIO}/vms/host1/ssh_port")
    local -r secondary_host_ssh_port=$(cat "${SCENARIO_INFO_DIR}/${SCENARIO}/vms/host2/ssh_port")
    local -r primary_host_name="$(full_vm_name host1)"
    local -r secondary_host_name="$(full_vm_name host2)"

    scp -P "${primary_host_ssh_port}" "${ROOTDIR}/scripts/multinode/configure-pri.sh" "redhat@${primary_host_ip}":
    ssh -p "${primary_host_ssh_port}" "redhat@${primary_host_ip}" ./configure-pri.sh "${primary_host_name}" "${primary_host_ip}" "${secondary_host_name}" "${secondary_host_ip}"

    scp -3 -P "${primary_host_ssh_port}" \
        "redhat@${primary_host_ip}:/home/redhat/kubelet-${secondary_host_name}.key" \
        "redhat@${primary_host_ip}:/home/redhat/kubelet-${secondary_host_name}.crt" \
        "redhat@${primary_host_ip}:/home/redhat/kubeconfig-${primary_host_name}" \
        "redhat@${secondary_host_ip}":

    scp -P "${secondary_host_ssh_port}" "${ROOTDIR}/scripts/multinode/configure-sec.sh" "redhat@${secondary_host_ip}":
    ssh -p "${secondary_host_ssh_port}" "redhat@${secondary_host_ip}" ./configure-sec.sh "${primary_host_name}" "${primary_host_ip}" "${secondary_host_name}" "${secondary_host_ip}"

    scp -P "${primary_host_ssh_port}" "redhat@${primary_host_ip}:/home/redhat/kubeconfig-${primary_host_name}" "${KUBECONFIG}"
    export KUBECONFIG="${KUBECONFIG}"
    echo "${primary_host_ip} ${primary_host_name}" | sudo tee -a /etc/hosts &>/dev/null
}

run_sonobuoy() {
    # Configure cluster prerequisites
    oc adm policy add-scc-to-group privileged system:authenticated system:serviceaccounts
    oc adm policy add-scc-to-group anyuid     system:authenticated system:serviceaccounts

    go install github.com/vmware-tanzu/sonobuoy@v0.56.16
    ~/go/bin/sonobuoy run \
        --mode=certified-conformance \
        --dns-namespace=openshift-dns \
        --dns-pod-labels=dns.operator.openshift.io/daemonset-dns=default 

    # Wait for up to 5m until tests start
    WAIT_FAILURE=true
    for _ in $(seq 1 150) ; do
        if [ "$(~/go/bin/sonobuoy status --json | jq -r '.status')" = "running" ]; then
            WAIT_FAILURE=false
            break
        fi
        sleep 2
    done
    if ${WAIT_FAILURE}; then
        echo "Failed to start tests after 5m"
        exit 1
    fi

    # Note that a normal run on 2 CPUs takes 40-45min.
    start=$(date +%s)
    while [ "$(~/go/bin/sonobuoy status --json | jq -r '.status')" = "running" ] ; do
        now=$(date +%s)
        if [ $(( now - start )) -ge ${TIMEOUT_TEST} ]; then
            echo "Tests running for ${TIMEOUT_TEST}s. Timing out"
            break
        fi
        ~/go/bin/sonobuoy status --json | jq '.plugins[] | select(.plugin=="e2e") | .progress'
        sleep 60
    done

    start=$(date +%s)
    # shellcheck disable=SC2046  # Jq is unable to process escaped quotes
    while [ -z $(~/go/bin/sonobuoy status --json | jq -r '."tar-info".name') ] ; do
        now=$(date +%s)
        if [ $(( now - start )) -ge ${TIMEOUT_RESULTS} ]; then
            echo "Waited for results for ${TIMEOUT_RESULTS}s. Timing out"
            break
        fi
        echo "Waiting for results availability"
        sleep 10
    done
    RESULTS_DIR=$(mktemp -d -p /tmp)
    ~/go/bin/sonobuoy retrieve "${RESULTS_DIR}" -f results.tar.gz
    tar xf "${RESULTS_DIR}/results.tar.gz" -C "${RESULTS_DIR}"
    cp "${RESULTS_DIR}/plugins/e2e/results/global/junit_01.xml" "${SCENARIO_INFO_DIR}/${SCENARIO}/"
    rm -r "${RESULTS_DIR}"

    if [ "$(~/go/bin/sonobuoy status --json | jq -r '.plugins[] | select(.plugin=="e2e") | .progress.failed')" != "null" ] ; then
        echo "Test was finished with errors. Check junit"
        exit 1
    fi
}

scenario_create_vms() {
    prepare_kickstart host1 kickstart.ks.template rhel-9.2-microshift-source
    prepare_kickstart host2 kickstart.ks.template rhel-9.2-microshift-source
    launch_vm host1
    launch_vm host2
}

scenario_remove_vms() {
    remove_vm host1
    remove_vm host2
}

scenario_run_tests() {
    prepare_hosts
    run_sonobuoy
}