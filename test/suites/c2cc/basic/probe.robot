*** Settings ***
Documentation       Verify C2CC probe pod deployment and health status reporting.
...                 Checks that the probe pod is deployed, the Service has the correct
...                 ClusterIP, RemoteCluster CRs transition to Healthy, and the
...                 deployment self-heals after deletion.

Resource            ../../../resources/microshift-process.resource
Resource            ../../../resources/kubeconfig.resource
Resource            ../../../resources/oc.resource
Resource            ../../../resources/c2cc.resource

Suite Setup         C2CC Suite Setup
Suite Teardown      C2CC Suite Teardown

Test Tags           c2cc


*** Variables ***
${C2CC_NAMESPACE}       openshift-c2cc
${PROBE_DEPLOYMENT}     c2cc-probe


*** Test Cases ***
Probe Namespace Exists
    [Documentation]    Verify the openshift-c2cc namespace exists on all clusters.
    FOR    ${alias}    IN    cluster-a    cluster-b    cluster-c
        ${stdout}=    Oc On Cluster    ${alias}    oc get namespace ${C2CC_NAMESPACE} -o name
        Should Contain    ${stdout}    namespace/${C2CC_NAMESPACE}
    END

Probe Deployment Running
    [Documentation]    Verify the c2cc-probe deployment is running with 1 ready replica.
    FOR    ${alias}    IN    cluster-a    cluster-b    cluster-c
        Wait Until Keyword Succeeds    2m    10s
        ...    Verify Probe Pod Is Ready    ${alias}
    END

Probe Service Has Correct ClusterIP
    [Documentation]    Verify the probe service has the 11th IP of the local service CIDR.
    Verify Probe Service ClusterIP    cluster-a    ${CLUSTER_A_SVC_CIDR}
    Verify Probe Service ClusterIP    cluster-b    ${CLUSTER_B_SVC_CIDR}
    Verify Probe Service ClusterIP    cluster-c    ${CLUSTER_C_SVC_CIDR}

RemoteCluster Status Becomes Healthy
    [Documentation]    Wait for RemoteCluster CRs to transition to Healthy on all clusters.
    Wait Until Keyword Succeeds    3m    10s
    ...    Verify RemoteCluster State    cluster-a    Healthy
    Wait Until Keyword Succeeds    3m    10s
    ...    Verify RemoteCluster State    cluster-b    Healthy
    Wait Until Keyword Succeeds    3m    10s
    ...    Verify RemoteCluster State    cluster-c    Healthy

RemoteCluster Status Has Probe Timestamps
    [Documentation]    Verify that LastProbeTime and LastSuccessfulProbe are populated
    ...    on all RemoteCluster CRs across all clusters.
    FOR    ${alias}    IN    cluster-a    cluster-b    cluster-c
        FOR    ${field}    IN    lastProbeTime    lastSuccessfulProbe
            ${stdout}=    Oc On Cluster    ${alias}
            ...    oc get remoteclusters.microshift.io -o jsonpath='{.items[*].status.${field}}'
            Should Not Be Empty    ${stdout}
            @{timestamps}=    Split String    ${stdout}
            ${count}=    Get Length    ${timestamps}
            Should Be Equal As Integers    ${count}    2    Expected 2 RemoteCluster ${field} values, got ${count}
            FOR    ${t}    IN    @{timestamps}
                Should Not Be Empty    ${t}
            END
        END
    END

RemoteCluster Status Has Latency Stats Per Target
    [Documentation]    Verify that all latency stat fields (avg/min/max/last/stddev) are populated
    ...    per target in TargetResults on all RemoteCluster CRs across all clusters.
    FOR    ${alias}    IN    cluster-a    cluster-b    cluster-c
        Wait Until Keyword Succeeds    2m    10s
        ...    Verify Per Target Latency Stats Populated    ${alias}
    END

Probe Deployment Self-Heals After Deletion
    [Documentation]    Delete the probe deployment and verify it is recreated by the controller.
    Oc On Cluster    cluster-a
    ...    oc delete deployment ${PROBE_DEPLOYMENT} -n ${C2CC_NAMESPACE}
    Wait Until Keyword Succeeds    2m    10s
    ...    Verify Probe Pod Is Ready    cluster-a

Probe Deployment Self-Heals After Scale Down
    [Documentation]    Scale down the probe deployment to 0 and verify it is restored to 1.
    Oc On Cluster    cluster-a
    ...    oc scale deployment ${PROBE_DEPLOYMENT} -n ${C2CC_NAMESPACE} --replicas=0
    Wait Until Keyword Succeeds    2m    10s
    ...    Verify Probe Pod Is Ready    cluster-a

RemoteCluster Status Becomes Unhealthy When Probe Fails
    [Documentation]    Block probe traffic on cluster-b and verify cluster-a
    ...    reports Unhealthy for the corresponding RemoteCluster CR.
    [Setup]    Verify All RemoteClusters Healthy
    ${primary_ip}=    Primary NextHop IP For Host    ${HOST2_IP}    ${HOST2_IPV6}
    ${cr_name}=    RemoteCluster CR Name From IP    ${primary_ip}
    # Apply a NetworkPolicy on cluster-b that denies all ingress to the probe pod,
    # causing cluster-a's probes to cluster-b to time out.
    Apply Probe Deny Policy    cluster-b
    # Wait for cluster-a to report Unhealthy (requires 3 consecutive failures)
    Wait Until Keyword Succeeds    3m    10s
    ...    Verify RemoteCluster State By Name    cluster-a    ${cr_name}    Unhealthy
    # Verify the Errors field is populated in the CR status
    ${errors}=    Get RemoteCluster Errors By Name    cluster-a    ${cr_name}
    Should Not Be Empty    ${errors}
    [Teardown]    Run Keywords
    ...    Delete Probe Deny Policy    cluster-b
    ...    AND    Wait Until Keyword Succeeds    3m    10s
    ...    Verify RemoteCluster State By Name    cluster-a    ${cr_name}    Healthy

Probe Service Has Correct Dual Stack ClusterIPs
    [Documentation]    In dual-stack, verify the probe service has secondary ClusterIPs.
    Skip If    '${CLUSTER_A_SVC_CIDR_DUAL}' == ''    Dual-stack CIDRs not configured
    Verify Probe Service ClusterIP Secondary    cluster-a    ${CLUSTER_A_SVC_CIDR_DUAL}
    Verify Probe Service ClusterIP Secondary    cluster-b    ${CLUSTER_B_SVC_CIDR_DUAL}
    Verify Probe Service ClusterIP Secondary    cluster-c    ${CLUSTER_C_SVC_CIDR_DUAL}

RemoteCluster Has TargetResults In Dual Stack
    [Documentation]    In dual-stack, verify TargetResults contains entries for both families.
    Skip If    '${CLUSTER_A_SVC_CIDR_DUAL}' == ''    Dual-stack CIDRs not configured
    FOR    ${alias}    IN    cluster-a    cluster-b    cluster-c
        Wait Until Keyword Succeeds    2m    10s
        ...    Verify TargetResults Populated    ${alias}
    END


*** Keywords ***
Verify Probe Pod Is Ready
    [Documentation]    Check that the probe deployment has 1 available replica.
    [Arguments]    ${alias}
    ${stdout}=    Oc On Cluster    ${alias}
    ...    oc get deployment ${PROBE_DEPLOYMENT} -n ${C2CC_NAMESPACE} -o jsonpath='{.status.availableReplicas}'
    Should Be Equal As Strings    ${stdout}    1

Verify Probe Service ClusterIP
    [Documentation]    Verify that the probe service ClusterIP matches the 11th IP of the given CIDR.
    [Arguments]    ${alias}    ${svc_cidr}
    ${expected_ip}=    Compute 11th IP    ${svc_cidr}
    ${actual_ip}=    Oc On Cluster    ${alias}
    ...    oc get service ${PROBE_DEPLOYMENT} -n ${C2CC_NAMESPACE} -o jsonpath='{.spec.clusterIP}'
    Should Be Equal As Strings    ${actual_ip}    ${expected_ip}    strip_spaces=True

Get RemoteCluster Errors By Name
    [Documentation]    Return the errors field from a specific RemoteCluster CR.
    [Arguments]    ${alias}    ${cr_name}
    ${stdout}=    Oc On Cluster
    ...    ${alias}
    ...    oc get remoteclusters.microshift.io ${cr_name} -o jsonpath='{.status.errors}'
    RETURN    ${stdout}

Apply Probe Deny Policy
    [Documentation]    Apply a NetworkPolicy that denies all ingress to the probe pod.
    [Arguments]    ${alias}
    ${policy}=    Catenate    SEPARATOR=\n
    ...    apiVersion: networking.k8s.io/v1
    ...    kind: NetworkPolicy
    ...    metadata:
    ...    ${SPACE}${SPACE}name: deny-probe-ingress
    ...    ${SPACE}${SPACE}namespace: ${C2CC_NAMESPACE}
    ...    spec:
    ...    ${SPACE}${SPACE}podSelector:
    ...    ${SPACE}${SPACE}${SPACE}${SPACE}matchLabels:
    ...    ${SPACE}${SPACE}${SPACE}${SPACE}${SPACE}${SPACE}app: c2cc-probe
    ...    ${SPACE}${SPACE}policyTypes:
    ...    ${SPACE}${SPACE}- Ingress
    Oc On Cluster    ${alias}    echo '${policy}' | oc apply -f -

Delete Probe Deny Policy
    [Documentation]    Remove the probe deny NetworkPolicy.
    [Arguments]    ${alias}
    Oc On Cluster    ${alias}
    ...    oc delete networkpolicy deny-probe-ingress -n ${C2CC_NAMESPACE} --ignore-not-found

Verify Per Target Latency Stats Populated
    [Documentation]    Check that all latency fields (avg/min/max/last/stddev) are populated
    ...    in TargetResults for all RemoteCluster CRs.
    [Arguments]    ${alias}
    FOR    ${field}    IN    avg    min    max    last    stddev
        ${stdout}=    Oc On Cluster    ${alias}
        ...    oc get remoteclusters.microshift.io -o jsonpath='{.items[*].status.targetResults[*].latency.${field}}'
        Should Not Be Empty    ${stdout}
        @{values}=    Split String    ${stdout}
        ${count}=    Get Length    ${values}
        # In single-stack: 2 remotes × 1 target = 2 values
        # In dual-stack: 2 remotes × 2 targets = 4 values
        Should Be True    ${count} >= 2    Expected at least 2 per-target latency ${field} values, got ${count}
        FOR    ${v}    IN    @{values}
            Should Not Be Empty    ${v}
        END
    END

Verify Probe Service ClusterIP Secondary
    [Documentation]    Verify that the probe service has a secondary ClusterIP matching the dual-stack CIDR.
    [Arguments]    ${alias}    ${svc_cidr_dual}
    ${expected_ip}=    Compute 11th IP    ${svc_cidr_dual}
    ${actual_ips}=    Oc On Cluster    ${alias}
    ...    oc get service ${PROBE_DEPLOYMENT} -n ${C2CC_NAMESPACE} -o jsonpath='{.spec.clusterIPs}'
    Should Contain    ${actual_ips}    ${expected_ip}

Verify TargetResults Populated
    [Documentation]    Verify that TargetResults contains per-family probe results.
    [Arguments]    ${alias}
    ${stdout}=    Oc On Cluster    ${alias}
    ...    oc get remoteclusters.microshift.io -o jsonpath='{.items[*].status.targetResults[*].target}'
    Should Not Be Empty    ${stdout}
    # In dual-stack with 2 remotes, we expect 4 total targets (2 families × 2 remotes)
    @{targets}=    Split String    ${stdout}
    ${count}=    Get Length    ${targets}
    Should Be True    ${count} >= 4    Expected at least 4 target results in dual-stack, got ${count}
