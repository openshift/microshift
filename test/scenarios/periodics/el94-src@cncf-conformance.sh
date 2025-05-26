#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

KUBECONFIG="${SCENARIO_INFO_DIR}/${SCENARIO}/kubeconfig"
# Timeout in seconds
TIMEOUT_TEST=7400
TIMEOUT_RESULTS=600

prepare_hosts() {
    local -r primary_host_ip=$(cat "${SCENARIO_INFO_DIR}/${SCENARIO}/vms/host1/ip")
    local -r secondary_host_ip=$(cat "${SCENARIO_INFO_DIR}/${SCENARIO}/vms/host2/ip")
    local -r primary_host_ssh_port=$(cat "${SCENARIO_INFO_DIR}/${SCENARIO}/vms/host1/ssh_port")
    local -r secondary_host_ssh_port=$(cat "${SCENARIO_INFO_DIR}/${SCENARIO}/vms/host2/ssh_port")
    local -r primary_host_name="$(full_vm_name host1)"
    local -r secondary_host_name="$(full_vm_name host2)"
    local rc=0

    # Configure primary host
    scp -P "${primary_host_ssh_port}" "${ROOTDIR}/scripts/multinode/configure-pri.sh" "redhat@${primary_host_ip}":
    ssh -p "${primary_host_ssh_port}" "redhat@${primary_host_ip}" ./configure-pri.sh \
        "${primary_host_name}" "${primary_host_ip}" \
        "${secondary_host_name}" "${secondary_host_ip}" || rc=$?
    if [ ${rc} -ne 0 ] ; then
        record_junit "prepare_hosts" "configure_primary" "FAILED"
        return ${rc}
    fi
    record_junit "prepare_hosts" "configure_primary" "OK"

    # Configure secondary host
    scp -3 -P "${primary_host_ssh_port}" \
        "redhat@${primary_host_ip}:/home/redhat/kubelet-${secondary_host_name}.key" \
        "redhat@${primary_host_ip}:/home/redhat/kubelet-${secondary_host_name}.crt" \
        "redhat@${primary_host_ip}:/home/redhat/kubeconfig-${primary_host_name}" \
        "redhat@${primary_host_ip}:/home/redhat/lvmd-${primary_host_name}.yaml" \
        "redhat@${secondary_host_ip}":

    scp -P "${secondary_host_ssh_port}" "${ROOTDIR}/scripts/multinode/configure-sec.sh" "redhat@${secondary_host_ip}":
    ssh -p "${secondary_host_ssh_port}" "redhat@${secondary_host_ip}" ./configure-sec.sh \
        "${primary_host_name}" "${primary_host_ip}" \
        "${secondary_host_name}" "${secondary_host_ip}" || rc=$?
    if [ ${rc} -ne 0 ] ; then
        record_junit "prepare_hosts" "configure_secondary" "FAILED"
        return ${rc}
    fi
    record_junit "prepare_hosts" "configure_secondary" "OK"

    # Configure kubeconfig and host name resolution
    scp -P "${primary_host_ssh_port}" "redhat@${primary_host_ip}:/home/redhat/kubeconfig-${primary_host_name}" "${KUBECONFIG}"
    export KUBECONFIG="${KUBECONFIG}"
    echo "${primary_host_ip} ${primary_host_name}" | sudo tee -a /etc/hosts &>/dev/null

    return ${rc}
}

run_sonobuoy() {
    local rc=0
    local start
    local now

    # Configure cluster prerequisites
    oc adm policy add-scc-to-group privileged system:authenticated system:serviceaccounts || rc=$?
    oc adm policy add-scc-to-group anyuid     system:authenticated system:serviceaccounts || rc=$?
    if [ ${rc} -ne 0 ] ; then
        record_junit "run_sonobuoy" "add_scc_to_group" "FAILED"
        return ${rc}
    fi

    # Initiate test startup
    go install github.com/vmware-tanzu/sonobuoy@v0.57.2
    ~/go/bin/sonobuoy run \
        --mode=certified-conformance \
        --plugin-env=e2e.E2E_SKIP=".*Services should be able to switch session affinity for NodePort service.*" \
        --dns-namespace=openshift-dns \
        --dns-pod-labels=dns.operator.openshift.io/daemonset-dns=default || rc=$?
    if [ ${rc} -ne 0 ] ; then
        record_junit "run_sonobuoy" "start_e2e" "FAILED"
        return ${rc}
    fi
    record_junit "run_sonobuoy" "start_e2e" "OK"

    # Wait for up to 5m until tests start
    rc=1
    for _ in $(seq 1 150) ; do
        if [ "$(~/go/bin/sonobuoy status --json | jq -r '.status')" = "running" ]; then
            rc=0
            break
        fi
        sleep 2
    done
    if [ ${rc} -ne 0 ] ; then
        echo "Failed to start tests after 5m"
        ~/go/bin/sonobuoy status --json
        record_junit "run_sonobuoy" "wait_e2e_running" "FAILED"
        return ${rc}
    fi
    record_junit "run_sonobuoy" "wait_e2e_running" "OK"

    # Note that a normal run on 2 CPUs takes 40-45min.
    local -r stat_file="${SCENARIO_INFO_DIR}/${SCENARIO}/cncf_status.json"
    start=$(date +%s)
    while true ; do
        ~/go/bin/sonobuoy status --json > "${stat_file}"
        cat "${stat_file}"
        if [ "$(jq -r '.status' "${stat_file}")" != "running" ] ; then
            break
        fi

        now=$(date +%s)
        if [ $(( now - start )) -ge ${TIMEOUT_TEST} ]; then
            rc=1
            echo "Tests running for ${TIMEOUT_TEST}s. Timing out"
            record_junit "run_sonobuoy" "wait_e2e_finished" "FAILED"
            break
        fi
        # Print progress information
        jq '.plugins[] | select(.plugin=="e2e") | .["result-counts"], .progress' "${stat_file}"
        sleep 60
    done
    # If the timeout is exceeded, proceed to attempt collecting test results
    if [ ${rc} -eq 0 ] ; then
        record_junit "run_sonobuoy" "wait_e2e_finished" "OK"
    fi

    local results=true
    start=$(date +%s)
    # shellcheck disable=SC2046  # Jq is unable to process escaped quotes
    while [ -z $(~/go/bin/sonobuoy status --json | jq -r '."tar-info".name') ] ; do
        now=$(date +%s)
        if [ $(( now - start )) -ge ${TIMEOUT_RESULTS} ]; then
            rc=1
            results=false
            echo "Waited for results for ${TIMEOUT_RESULTS}s. Timing out"
            record_junit "run_sonobuoy" "wait_e2e_results" "FAILED"
            break
        fi
        echo "Waiting for results availability"
        sleep 10
    done

    if ${results} ; then
        record_junit "run_sonobuoy" "wait_e2e_results" "OK"

        # Collect the results
        local -r results_dir=$(mktemp -d -p /tmp)
        ~/go/bin/sonobuoy retrieve "${results_dir}" -f results.tar.gz
        tar xf "${results_dir}/results.tar.gz" -C "${results_dir}"
        cp "${results_dir}/results.tar.gz" "${SCENARIO_INFO_DIR}/${SCENARIO}/sonobuoy-results.tar.gz"
        cp "${results_dir}/plugins/e2e/results/global/"{e2e.log,junit_01.xml} "${SCENARIO_INFO_DIR}/${SCENARIO}/"
        rm -r "${results_dir}"

        # If we got the results we need to check if there are any failures
        # Failures without logs are useless
        local -r failures=$(~/go/bin/sonobuoy status --json | jq '[.plugins[] | select(."result-status" == "failed")] | length')
        if [ "${failures}" != "0" ]; then
            rc=1
        fi
    fi

    if [ ${rc} -eq 0 ] ; then
        echo "Tests finished successfully"
        record_junit "run_sonobuoy" "run_e2e_status" "OK"
    else
        echo "Tests finished with errors. Check junit"
        record_junit "run_sonobuoy" "run_e2e_status" "FAILED"
    fi
    return ${rc}
}

scenario_create_vms() {
    prepare_kickstart host1 kickstart.ks.template rhel-9.4-microshift-source
    prepare_kickstart host2 kickstart.ks.template rhel-9.4-microshift-source
    launch_vm --vmname host1 --boot_blueprint rhel-9.4
    launch_vm --vmname host2 --boot_blueprint rhel-9.4
}

scenario_remove_vms() {
    remove_vm host1
    remove_vm host2
}

scenario_run_tests() {
    if ! prepare_hosts ; then
        return 1
    fi
    if ! run_sonobuoy ; then
        return 1
    fi
}
