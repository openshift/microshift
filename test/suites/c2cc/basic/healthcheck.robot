*** Settings ***
Documentation       Verify C2CC RemoteCluster CRD and CR lifecycle.
...                 Checks that the CRD is registered, CRs are created per remote cluster,
...                 and CR specs match the expected probe targets.

Resource            ../../../resources/microshift-process.resource
Resource            ../../../resources/kubeconfig.resource
Resource            ../../../resources/oc.resource
Resource            ../../../resources/c2cc.resource

Suite Setup         C2CC Suite Setup
Suite Teardown      C2CC Suite Teardown

Test Tags           c2cc


*** Test Cases ***
RemoteCluster CRD Exists
    [Documentation]    Verify the remoteclusters.microshift.io CRD is registered on all clusters.
    [Template]    Wait For RemoteCluster CRD
    cluster-a
    cluster-b
    cluster-c

RemoteCluster CR Created
    [Documentation]    Verify exactly one RemoteCluster CR exists on all clusters.
    [Template]    Wait For Correct RemoteCluster CR Count
    cluster-a    2
    cluster-b    2
    cluster-c    2

Correct RemoteCluster CR Spec
    [Documentation]    Verify each RemoteCluster CR has the correct probe target and interval.
    [Template]    Verify RemoteCluster CR Spec
    cluster-a    ${CLUSTER_B_SVC_CIDR}
    cluster-a    ${CLUSTER_C_SVC_CIDR}
    cluster-b    ${CLUSTER_A_SVC_CIDR}
    cluster-b    ${CLUSTER_C_SVC_CIDR}
    cluster-c    ${CLUSTER_A_SVC_CIDR}
    cluster-c    ${CLUSTER_B_SVC_CIDR}

RemoteCluster CR Has Managed-By Label
    [Documentation]    Verify all RemoteCluster CRs have the expected managed-by label.
    [Template]    Verify RemoteCluster CR Label
    cluster-a
    cluster-b
    cluster-c

Correct Dual Stack RemoteCluster CR Spec
    [Documentation]    In dual-stack, verify probe targets contain both address families.
    Skip If    '${CLUSTER_A_SVC_CIDR_DUAL}' == ''    Dual-stack CIDRs not configured
    Verify RemoteCluster CR Spec    cluster-a    ${CLUSTER_B_SVC_CIDR_DUAL}
    Verify RemoteCluster CR Spec    cluster-a    ${CLUSTER_C_SVC_CIDR_DUAL}
    Verify RemoteCluster CR Spec    cluster-b    ${CLUSTER_A_SVC_CIDR_DUAL}
    Verify RemoteCluster CR Spec    cluster-b    ${CLUSTER_C_SVC_CIDR_DUAL}
    Verify RemoteCluster CR Spec    cluster-c    ${CLUSTER_A_SVC_CIDR_DUAL}
    Verify RemoteCluster CR Spec    cluster-c    ${CLUSTER_B_SVC_CIDR_DUAL}


*** Keywords ***
Wait For RemoteCluster CRD
    [Documentation]    Waits for remoteclusters.microshift.io CRD to be registered.
    [Arguments]    ${alias}
    Wait Until Keyword Succeeds    2m    15s
    ...    Verify RemoteCluster CRD Exists    ${alias}

Verify RemoteCluster CRD Exists
    [Documentation]    Verify that the remoteclusters.microshift.io CRD is registered.
    [Arguments]    ${alias}
    ${stdout}=    Oc On Cluster    ${alias}    oc get crd remoteclusters.microshift.io -o name
    Should Be Equal As Strings
    ...    ${stdout}
    ...    customresourcedefinition.apiextensions.k8s.io/remoteclusters.microshift.io
    ...    strip_spaces=True

Wait For Correct RemoteCluster CR Count
    [Documentation]    Verify the number of RemoteCluster CRs matches the expected count.
    [Arguments]    ${alias}    ${expected_count}
    Wait Until Keyword Succeeds    2m    15s
    ...    Verify RemoteCluster CR Count    ${alias}    ${expected_count}

Verify RemoteCluster CR Count
    [Documentation]    Verify the number of RemoteCluster CRs matches the expected count.
    [Arguments]    ${alias}    ${expected_count}
    ${stdout}=    Oc On Cluster    ${alias}
    ...    oc get remoteclusters.microshift.io -l app.kubernetes.io/managed-by=c2cc-route-manager -o name
    @{lines}=    Split To Lines    ${stdout}
    ${count}=    Get Length    ${lines}
    Should Be Equal As Integers    ${count}    ${expected_count}

Verify RemoteCluster CR Spec
    [Documentation]    Verify that a RemoteCluster CR exists with the correct probe target
    ...    (11th IP in the remote service CIDR on port 8080)
    ...    and a non-empty probe interval duration string.
    [Arguments]    ${alias}    ${remote_svc_cidr}
    ${expected_ip}=    Compute 11th IP    ${remote_svc_cidr}
    IF    ":" in """${expected_ip}"""
        VAR    ${expected_target}=    [${expected_ip}]:8080
    ELSE
        VAR    ${expected_target}=    ${expected_ip}:8080
    END
    ${targets}=    Oc On Cluster
    ...    ${alias}
    ...    oc get remoteclusters.microshift.io -l app.kubernetes.io/managed-by=c2cc-route-manager -o jsonpath='{.items[*].spec.probeTargets[*]}'
    Should Contain    ${targets}    ${expected_target}
    ${intervals}=    Oc On Cluster
    ...    ${alias}
    ...    oc get remoteclusters.microshift.io -l app.kubernetes.io/managed-by=c2cc-route-manager -o jsonpath='{.items[*].spec.probeInterval}'
    Should Not Be Empty    ${intervals}
    @{interval_list}=    Split String    ${intervals}
    FOR    ${interval}    IN    @{interval_list}
        Should Match Regexp    ${interval}    ^[0-9]+(s|m|h)$
    END

Verify RemoteCluster CR Label
    [Documentation]    Verify all RemoteCluster CRs have the app.kubernetes.io/managed-by=c2cc-route-manager label.
    [Arguments]    ${alias}
    ${stdout}=    Oc On Cluster
    ...    ${alias}
    ...    oc get remoteclusters.microshift.io -l app.kubernetes.io/managed-by=c2cc-route-manager -o jsonpath='{.items[*].metadata.labels.app\\.kubernetes\\.io/managed-by}'
    Should Not Be Empty    ${stdout}
    @{labels}=    Split String    ${stdout}
    FOR    ${label}    IN    @{labels}
        Should Be Equal As Strings    ${label}    c2cc-route-manager    strip_spaces=True
    END
