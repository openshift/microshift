*** Settings ***
Documentation       Sanity checks for a multi-cluster C2CC environment.
...                 Verifies that all clusters are running and MicroShift is healthy.

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-host.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/kubeconfig.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/c2cc.resource

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           c2cc


*** Test Cases ***
Cluster A Is Running
    [Documentation]    Verify MicroShift on Cluster A is healthy.
    Verify Cluster Is Running    cluster-a

Cluster B Is Running
    [Documentation]    Verify MicroShift on Cluster B is healthy.
    Verify Cluster Is Running    cluster-b

All Pods On Cluster A Are Ready
    [Documentation]    All pods on Cluster A reach Ready state.
    Verify All Pods Are Ready    cluster-a

All Pods On Cluster B Are Ready
    [Documentation]    All pods on Cluster B reach Ready state.
    Verify All Pods Are Ready    cluster-b

Cluster A Has Expected Node
    [Documentation]    Verify Cluster A has a node.
    Verify Cluster Has Node    cluster-a

Cluster B Has Expected Node
    [Documentation]    Verify Cluster B has a node.
    Verify Cluster Has Node    cluster-b

C2CC Controller Is Running On Cluster A
    [Documentation]    Verify c2cc-route-manager logged startup on Cluster A.
    Verify C2CC Controller Is Running    cluster-a

C2CC Controller Is Running On Cluster B
    [Documentation]    Verify c2cc-route-manager logged startup on Cluster B.
    Verify C2CC Controller Is Running    cluster-b


*** Keywords ***
Setup
    [Documentation]    Set up SSH connections and kubeconfigs for all clusters.
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig
    Register Local Cluster    cluster-a
    Register Remote Cluster    cluster-b    ${HOST2_IP}    ${HOST2_SSH_PORT}    ${KUBECONFIG_B}

Teardown
    [Documentation]    Close all connections and clean up kubeconfigs.
    Teardown All Remote Clusters
    Remove Kubeconfig
    Logout MicroShift Host

Verify Cluster Is Running
    [Documentation]    Check the /readyz endpoint on the given cluster.
    [Arguments]    ${alias}
    ${stdout}=    Oc On Cluster    ${alias}    oc get --raw='/readyz'
    Should Be Equal As Strings    ${stdout}    ok    strip_spaces=True

Verify All Pods Are Ready
    [Documentation]    Wait for all pods to be Ready on the given cluster.
    [Arguments]    ${alias}
    Oc On Cluster    ${alias}    oc wait pods -A --all --for=condition=Ready --timeout=120s

Verify Cluster Has Node
    [Documentation]    Verify the given cluster has at least one node.
    [Arguments]    ${alias}
    ${output}=    Oc On Cluster    ${alias}    oc get nodes -o name
    Should Not Be Empty    ${output}

Verify C2CC Controller Is Running
    [Documentation]    Verify c2cc-route-manager logged startup in the journal.
    [Arguments]    ${alias}
    ${stdout}=    Command On Cluster    ${alias}    journalctl -u microshift --grep "C2CC is enabled" --no-pager -q
    Should Contain    ${stdout}    C2CC is enabled
