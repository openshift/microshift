*** Settings ***
Documentation       Verify C2CC probe pod deployment and health status reporting.
...                 Checks that the probe pod is deployed, the Service has the correct
...                 ClusterIP, RemoteCluster CRs transition to Healthy, and the
...                 deployment self-heals after deletion.

Resource            ../../resources/microshift-process.resource
Resource            ../../resources/kubeconfig.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/c2cc.resource

Suite Setup         Setup
Suite Teardown      Teardown

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

RemoteCluster Status Has LastProbeTime
    [Documentation]    Verify that LastProbeTime is populated on all RemoteCluster CRs.
    FOR    ${alias}    IN    cluster-a    cluster-b    cluster-c
        ${stdout}=    Oc On Cluster    ${alias}
        ...    oc get remoteclusters.microshift.io -o jsonpath='{.items[*].status.lastProbeTime}'
        Should Not Be Empty    ${stdout}
        @{timestamps}=    Split String    ${stdout}
        ${count}=    Get Length    ${timestamps}
        Should Be Equal As Integers    ${count}    2    Expected 2 RemoteCluster states, got ${count}
        FOR    ${t}    IN    @{timestamps}
            Should Not Be Empty    ${t}
        END
    END

RemoteCluster Status Has LastSuccessfulProbe
    [Documentation]    Verify that LastSuccessfulProbe is populated on all RemoteCluster CRs.
    FOR    ${alias}    IN    cluster-a    cluster-b    cluster-c
        ${stdout}=    Oc On Cluster    ${alias}
        ...    oc get remoteclusters.microshift.io -o jsonpath='{.items[*].status.lastSuccessfulProbe}'
        Should Not Be Empty    ${stdout}
        @{timestamps}=    Split String    ${stdout}
        ${count}=    Get Length    ${timestamps}
        Should Be Equal As Integers    ${count}    2    Expected 2 RemoteCluster states, got ${count}
        FOR    ${t}    IN    @{timestamps}
            Should Not Be Empty    ${t}
        END
    END

RemoteCluster Status Has Latency Stats
    [Documentation]    Verify that latency statistics are populated after probes have run.
    FOR    ${alias}    IN    cluster-a    cluster-b
        Wait Until Keyword Succeeds    2m    10s
        ...    Verify Latency Stats Populated    ${alias}
    END

Latency Stats Fields Are Populated
    [Documentation]    Verify all latency stat fields (avg/min/max/last/stddev) are present and non-empty.
    FOR    ${alias}    IN    cluster-a    cluster-b
        ${avg}=    Get Latency Field    ${alias}    avg
        ${min}=    Get Latency Field    ${alias}    min
        ${max}=    Get Latency Field    ${alias}    max
        ${last}=    Get Latency Field    ${alias}    last
        ${stddev}=    Get Latency Field    ${alias}    stddev
        Should Not Be Empty    ${avg}
        Should Not Be Empty    ${min}
        Should Not Be Empty    ${max}
        Should Not Be Empty    ${last}
        Should Not Be Empty    ${stddev}
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
    [Setup]    Ensure All Clusters Healthy
    ${cr_name}=    RemoteCluster CR Name From IP    ${HOST2_IP}
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


*** Keywords ***
Setup
    [Documentation]    Set up SSH connections and kubeconfigs for all clusters.
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig
    Register Local Cluster    cluster-a
    Register Remote Cluster    cluster-b    ${HOST2_IP}    ${HOST2_SSH_PORT}    ${KUBECONFIG_B}
    Register Remote Cluster    cluster-c    ${HOST3_IP}    ${HOST3_SSH_PORT}    ${KUBECONFIG_C}

Teardown
    [Documentation]    Close all connections and clean up kubeconfigs.
    Teardown All Remote Clusters
    Remove Kubeconfig
    Logout MicroShift Host

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

Verify RemoteCluster State
    [Documentation]    Check that all RemoteCluster CRs on this cluster have the expected state.
    [Arguments]    ${alias}    ${expected_state}
    ${stdout}=    Oc On Cluster    ${alias}
    ...    oc get remoteclusters.microshift.io -o jsonpath='{.items[*].status.state}'
    Should Not Be Empty    ${stdout}
    @{states}=    Split String    ${stdout}
    ${count}=    Get Length    ${states}
    Should Be Equal As Integers    ${count}    2    Expected 2 RemoteCluster states, got ${count}
    FOR    ${state}    IN    @{states}
        Should Be Equal As Strings    ${state}    ${expected_state}
    END

Verify RemoteCluster State By Name
    [Documentation]    Check that a specific RemoteCluster CR has the expected state.
    [Arguments]    ${alias}    ${cr_name}    ${expected_state}
    ${stdout}=    Oc On Cluster
    ...    ${alias}
    ...    oc get remoteclusters.microshift.io ${cr_name} -o jsonpath='{.status.state}'
    Should Be Equal As Strings    ${stdout}    ${expected_state}

Get RemoteCluster Errors By Name
    [Documentation]    Return the errors field from a specific RemoteCluster CR.
    [Arguments]    ${alias}    ${cr_name}
    ${stdout}=    Oc On Cluster
    ...    ${alias}
    ...    oc get remoteclusters.microshift.io ${cr_name} -o jsonpath='{.status.errors}'
    RETURN    ${stdout}

RemoteCluster CR Name From IP
    [Documentation]    Compute the RemoteCluster CR name from a host IP (e.g. 192.168.1.2 -> c2cc-192-168-1-2).
    [Arguments]    ${ip}
    ${dashed}=    Replace String    ${ip}    .    -
    ${dashed}=    Replace String    ${dashed}    :    -
    RETURN    c2cc-${dashed}

Ensure All Clusters Healthy
    [Documentation]    Pre-condition: all clusters must be Healthy before fault injection.
    FOR    ${alias}    IN    cluster-a    cluster-b    cluster-c
        Wait Until Keyword Succeeds    3m    10s
        ...    Verify RemoteCluster State    ${alias}    Healthy
    END

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

Verify Latency Stats Populated
    [Documentation]    Check that latency stats are present and avg is non-empty.
    [Arguments]    ${alias}
    ${avg}=    Get Latency Field    ${alias}    avg
    Should Not Be Empty    ${avg}

Get Latency Field
    [Documentation]    Return a single latency stat field from the first RemoteCluster CR.
    [Arguments]    ${alias}    ${field}
    ${stdout}=    Oc On Cluster    ${alias}
    ...    oc get remoteclusters.microshift.io -o jsonpath='{.items[0].status.latency.${field}}'
    RETURN    ${stdout}
