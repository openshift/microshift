*** Settings ***
Documentation       Sanity checks for a multi-cluster C2CC environment.
...                 Verifies that all clusters are running and MicroShift is healthy.

Resource            ../../../resources/microshift-process.resource
Resource            ../../../resources/kubeconfig.resource
Resource            ../../../resources/oc.resource
Resource            ../../../resources/c2cc.resource

Suite Setup         C2CC Suite Setup
Suite Teardown      C2CC Suite Teardown

Test Tags           c2cc


*** Test Cases ***
Cluster Is Running
    [Documentation]    Verify all clusters are healthy.
    [Template]    Verify Cluster Is Running
    cluster-a
    cluster-b
    cluster-c

All Pods On Cluster Are Ready
    [Documentation]    Verify all pods reach Ready state on all clusters.
    [Template]    Verify All Pods Are Ready
    cluster-a
    cluster-b
    cluster-c

Cluster Has Expected Node
    [Documentation]    Verify all clusters have a node.
    [Template]    Verify Cluster Has Node
    cluster-a
    cluster-b
    cluster-c

C2CC Controller Is Running On Cluster
    [Documentation]    Verify c2cc-route-manager logged startup on all clusters.
    [Template]    Verify C2CC Controller Is Running
    cluster-a
    cluster-b
    cluster-c


*** Keywords ***
Verify Cluster Is Running
    [Documentation]    Check the /readyz endpoint on the given cluster.
    [Arguments]    ${alias}
    ${stdout}=    Oc On Cluster    ${alias}    oc get --raw='/readyz'
    Should Be Equal As Strings    ${stdout}    ok    strip_spaces=True

Verify All Pods Are Ready
    [Documentation]    Wait for all pods to be Ready on the given cluster.
    [Arguments]    ${alias}
    Oc On Cluster
    ...    ${alias}
    ...    oc wait pods -A --all --for=condition=Ready --field-selector=status.phase!=Succeeded,status.phase!=Failed --timeout=120s

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
