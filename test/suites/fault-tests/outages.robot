*** Settings ***
Documentation       Fault Test For MicroShift

Library             OperatingSystem
Library             Process
Library             Collections
Library             yaml
Library             ../../resources/journalctl.py
Resource            ../../resources/common.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/microshift-process.resource

Suite Setup         Setup
Suite Teardown      Teardown


*** Variables ***
${USHIFT_HOST}                      ${EMPTY}
${USHIFT_USER}                      ${EMPTY}
${SSH_PORT}                         ${EMPTY}
${SSH_PRIV_KEY}                     ${EMPTY}
${STRESS_TESTING_SCRIPT}            bin/stress_testing.sh
${STRESS_TESTING_REMOTE_FLAGS}      -h ${USHIFT_HOST} -u ${USHIFT_USER} -p ${SSH_PORT} -k ${SSH_PRIV_KEY}


*** Test Cases ***
Network Disconnection
    [Documentation]    Test system behavior during network disconnection
    # Get init values
    ${cursor}=    Get Journal Cursor
    ${old_bootid}=    Get Current Boot Id

    # Run scenario
    Local Command Should Work    ${STRESS_TESTING_SCRIPT} -e network_outage ${STRESS_TESTING_REMOTE_FLAGS}
    Sleep    15s
    ${stdout}    ${rc}=    Run With Kubeconfig    timeout 10s oc get nodes    allow_fail=True    return_rc=True
    Should Be Equal As Numbers    ${rc}    124
    Sleep    15s
    Local Command Should Work    ${STRESS_TESTING_SCRIPT} -d network_outage ${STRESS_TESTING_REMOTE_FLAGS}

    # Check results
    ${system_rebooted}=    Is System Rebooted    ${old_bootid}
    Should Not Be True    ${system_rebooted}
    Wait For MicroShift

    ${output}    ${rc}=    Get Log Output With Pattern    ${cursor}    kubelet
    ${expected_str}=    Get Expected Message    suites/fault-tests/log-messages.yaml    disconnect    network
    Compare Output Logs    ${output}    ${expected_str}

Delete A Pod
    [Documentation]    Test to delete pods and verify they are recreated properly
    [Template]    Delete in ${NAMESPACE} ${POD_NAME} pod
    kube-system    csi-snapshot-controller
    openshift-dns    dns-default
    openshift-dns    node-resolver
    openshift-ingress    router-default
    openshift-ovn-kubernetes    ovnkube-master
    openshift-ovn-kubernetes    ovnkube-node
    openshift-service-ca    service-ca
    openshift-storage    lvms-operator
    openshift-storage    vg-manager


*** Keywords ***
Setup
    [Documentation]    Set up all of the tests in this suite
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig

Teardown
    [Documentation]    Test suite teardown
    Remove Kubeconfig
    Logout MicroShift Host

Delete In ${namespace} ${pod_name} Pod
    [Documentation]    Delete a pod in the specified namespace and verify the expected log messages
    ${cursor}=    Get Journal Cursor
    Delete Pod And Wait For Recovery    ${namespace}    ${pod_name}

    ${actual_str}    ${rc}=    Get Log Output With Pattern    ${cursor}    kubelet.go.*${namespace}
    ${expected_str}=    Get Expected Message    suites/fault-tests/log-messages.yaml    delete    pod
    ${expected_str}=    Replace String    ${expected_str}    {{namespace}}    ${namespace}
    ${expected_str}=    Replace String    ${expected_str}    {{pod_name}}    ${pod_name}
    Compare Output Logs    ${actual_str}    ${expected_str}

Delete Pod And Wait For Recovery
    [Documentation]    Delete a pod and wait for it to be recreated and ready
    [Arguments]    ${namespace}    ${pod_name}
    Run With Kubeconfig
    ...    oc get pod -n ${namespace} | grep ${pod_name} | awk '{print $1}' | xargs oc delete pod -n ${namespace} --force --grace-period=0
    ...    allow_fail=${TRUE}
    Sleep    5s
    All Pods Should Be Running    ns=${namespace}
