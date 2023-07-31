#!/bin/bash

# Sourced from cleanup_scenario.sh and uses functions defined there.

prepare_hosts() {
    local primary_host_ip=$(cat "${SCENARIO_INFO_DIR}/${SCENARIO}/vms/host1/public_ip")
    local secondary_host_ip=$(cat "${SCENARIO_INFO_DIR}/${SCENARIO}/vms/host2/public_ip")
    #tOdo need the ports too
    local primary_host_name="$(full_vm_name host1)"
    local secondary_host_name="$(full_vm_name host2)"

    scp "${ROOTDIR}/scripts/multinode/configure-pri.sh" "redhat@${primary_host_ip}":
    ssh redhat@${primary_host_ip} ./configure-pri.sh "${primary_host_name}" "${primary_host_ip}" "${secondary_host_name}" "${secondary_host_ip}"

    scp -3 \
        redhat@${primary_host_ip}:/home/redhat/kubelet-${secondary_host_name}.{key,crt} \
        redhat@${primary_host_ip}:/home/redhat/kubeconfig-${primary_host_name} \
        redhat@${secondary_host_ip}:

    scp "${ROOTDIR}/scripts/multinode/configure-sec.sh" redhat@${secondary_host_ip}:
    ssh redhat@${secondary_host_ip} ./configure-sec.sh "${primary_host_name}" "${primary_host_ip}" "${secondary_host_name}" "${secondary_host_ip}"

    export KUBECONFIG=$(mktemp /tmp/microshift-kubeconfig.XXXXXXXXXX)
    scp redhat@${primary_host_ip}:/home/redhat/kubeconfig-${primary_host_name} ${KUBECONFIG}
    echo "${primary_host_ip} ${primary_host_name}" | sudo tee -a /etc/hosts &>/dev/null
}

run_sonobuoy() {
    # Configure cluster prerequisites
    oc adm policy add-scc-to-group privileged system:authenticated system:serviceaccounts
    oc adm policy add-scc-to-group anyuid     system:authenticated system:serviceaccounts

    go install github.com/vmware-tanzu/sonobuoy@latest
    ~/go/bin/sonobuoy run \
     --mode=certified-conformance \
     --dns-namespace=openshift-dns \
     --dns-pod-labels=dns.operator.openshift.io/daemonset-dns=default

    # Wait for up to 1m until tests start
    WAIT_FAILURE=true
    for _ in $(seq 1 30) ; do
        if ~/go/bin/sonobuoy status --json | jq '.status' &>/dev/null ; then
            WAIT_FAILURE=false
            break
        fi
        sleep 2
    done
    ${WAIT_FAILURE} && exit 1

    # Wait until test complete (exit as soon as one of the tests failed)
    TEST_FAILURE=false
    while [ "$(~/go/bin/sonobuoy status --json | jq -r '.status')" = "running" ] ; do
        ~/go/bin/sonobuoy status --json | jq '.plugins[] | select(.plugin=="e2e") | .progress'
        if [ "$(~/go/bin/sonobuoy status --json | jq -r '.plugins[] | select(.plugin=="e2e") | .progress.failed')" != "null" ] ; then
            TEST_FAILURE=true
            break
        fi
        sleep 60
    done
    ${TEST_FAILURE} && exit 1
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