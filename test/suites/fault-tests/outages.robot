*** Settings ***
Documentation       Fault Test For MicroShift

Library             OperatingSystem
Library             Process
Library             Collections
Library             yaml
Library             ../../resources/journalctl.py
Resource            ../../resources/fault-tests.resource
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
    ${stdout}    ${rc}=    Run With Kubeconfig    oc get nodes    allow_fail=True    return_rc=True    timeout=30s
    Should Be Equal As Numbers    ${rc}    -15    # -15 is the SIGTERM signal because of timeout in previous command
    Sleep    15s
    Local Command Should Work    ${STRESS_TESTING_SCRIPT} -d network_outage ${STRESS_TESTING_REMOTE_FLAGS}

    # Check results
    ${system_rebooted}=    Is System Rebooted    ${old_bootid}
    Should Not Be True    ${system_rebooted}
    Wait For MicroShift
    All Pods Should Be Running    timeout=600s

    ${output}    ${rc}=    Get Log Output With Pattern    ${cursor}    kubelet
    @{expected_lines}=    Get Expected Messages    suites/fault-tests/log-messages.yaml    disconnect    network
    Compare Output Logs    ${output}    @{expected_lines}

Delete A Pod
    [Documentation]    Test to delete pods and verify they are recreated properly
    [Template]    Delete Pod ${POD_NAME} In ${NAMESPACE}
    csi-snapshot-controller    kube-system
    dns-default    openshift-dns
    node-resolver    openshift-dns
    router-default    openshift-ingress
    ovnkube-master    openshift-ovn-kubernetes
    ovnkube-node    openshift-ovn-kubernetes
    service-ca    openshift-service-ca
    lvms-operator    openshift-storage
    vg-manager    openshift-storage


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

Delete Pod ${pod_name} In ${namespace}
    [Documentation]    Delete a pod in the specified namespace and verify the expected log messages
    ${cursor}=    Get Journal Cursor
    Delete Pod And Wait For Recovery    ${namespace}    ${pod_name}

    ${actual_str}    ${rc}=    Get Log Output With Pattern    ${cursor}    kubelet.go.*${namespace}
    @{expected_lines}=    Get Expected Messages    suites/fault-tests/log-messages.yaml    delete    pod
    ${expected_lines_replaced}=    Create List
    FOR    ${line}    IN    @{expected_lines}
        ${line_with_vars}=    Replace Variables    ${line}
        Append To List    ${expected_lines_replaced}    ${line_with_vars}
    END
    Compare Output Logs    ${actual_str}    @{expected_lines_replaced}

Delete Pod And Wait For Recovery
    [Documentation]    Delete a pod and wait for it to be recreated and ready
    [Arguments]    ${namespace}    ${pod_name}
    Run With Kubeconfig
    ...    oc get pod -n ${namespace} -o=name | grep ${pod_name} | xargs oc delete -n ${namespace} --force --grace-period=0
    ...    allow_fail=${TRUE}
    Sleep    5s
    All Pods Should Be Running    ns=${namespace}
