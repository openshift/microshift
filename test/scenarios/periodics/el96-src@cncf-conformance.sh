#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

KUBECONFIG="${SCENARIO_INFO_DIR}/${SCENARIO}/kubeconfig"
# Timeout in seconds
TIMEOUT_TEST=7400
TIMEOUT_RESULTS=600

collect_sonobuoy_debug_info() {
    ~/go/bin/sonobuoy logs > "${SCENARIO_INFO_DIR}/${SCENARIO}/sonobuoy-logs.txt" || true
    oc get all -n sonobuoy -o wide > "${SCENARIO_INFO_DIR}/${SCENARIO}/sonobuoy-resources.txt" || true
    oc describe all -n sonobuoy > "${SCENARIO_INFO_DIR}/${SCENARIO}/sonobuoy-resources-describe.txt" || true
    oc get events -n sonobuoy --sort-by=.metadata.creationTimestamp > "${SCENARIO_INFO_DIR}/${SCENARIO}/sonobuoy-events.txt" || true
}

prepare_hosts() {
    local -r primary_host_ip=$(cat "${SCENARIO_INFO_DIR}/${SCENARIO}/vms/host1/ip")
    local -r secondary_host_ip=$(cat "${SCENARIO_INFO_DIR}/${SCENARIO}/vms/host2/ip")
    local -r primary_host_ssh_port=$(cat "${SCENARIO_INFO_DIR}/${SCENARIO}/vms/host1/ssh_port")
    local -r secondary_host_ssh_port=$(cat "${SCENARIO_INFO_DIR}/${SCENARIO}/vms/host2/ssh_port")
    local rc=0

    # Configure primary host
    scp -P "${primary_host_ssh_port}" "${ROOTDIR}/scripts/multinode/configure-node.sh" "redhat@${primary_host_ip}":
    ssh -p "${primary_host_ssh_port}" "redhat@${primary_host_ip}" ./configure-node.sh || rc=$?
    if [ ${rc} -ne 0 ] ; then
        record_junit "prepare_hosts" "configure_primary" "FAILED"
        return ${rc}
    fi
    record_junit "prepare_hosts" "configure_primary" "OK"

    # Configure secondary host
    scp -P "${primary_host_ssh_port}" \
        "redhat@${primary_host_ip}:/home/redhat/kubeconfig-bootstrap" \
        "redhat@${secondary_host_ip}":

    scp -P "${secondary_host_ssh_port}" "${ROOTDIR}/scripts/multinode/configure-node.sh" "redhat@${secondary_host_ip}":
    ssh -p "${secondary_host_ssh_port}" "redhat@${secondary_host_ip}" "./configure-node.sh --bootstrap-kubeconfig /home/redhat/kubeconfig-bootstrap" || rc=$?
    if [ ${rc} -ne 0 ] ; then
        record_junit "prepare_hosts" "configure_secondary" "FAILED"
        return ${rc}
    fi
    record_junit "prepare_hosts" "configure_secondary" "OK"

    scp -P "${primary_host_ssh_port}" "redhat@${primary_host_ip}:/home/redhat/kubeconfig-bootstrap" "${KUBECONFIG}"
    export KUBECONFIG="${KUBECONFIG}"

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
    go install "github.com/vmware-tanzu/sonobuoy@${CNCF_SONOBUOY_VERSION}"
    # Force the images to include the registry, as default values may introduce ambiguity (e.g. sonobuoy/sonobuoy:v0.57.3)
    ~/go/bin/sonobuoy run \
	    --sonobuoy-image "docker.io/sonobuoy/sonobuoy:${CNCF_SONOBUOY_VERSION}" \
        --systemd-logs-image "docker.io/sonobuoy/systemd-logs:${CNCF_SYSTEMD_LOGS_VERSION}" \
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
        # Retrieve sonobuoy info in case the tests have not started yet. This covers scheduling/container creation issues.
        collect_sonobuoy_debug_info
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
            collect_sonobuoy_debug_info
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
    prepare_kickstart host1 kickstart.ks.template rhel-9.6-microshift-source
    prepare_kickstart host2 kickstart.ks.template rhel-9.6-microshift-source
    launch_vm --vmname host1 --boot_blueprint rhel-9.6
    launch_vm --vmname host2 --boot_blueprint rhel-9.6
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
