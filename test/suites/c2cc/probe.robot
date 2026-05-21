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
${C2CC_NAMESPACE}       microshift-c2cc
${PROBE_DEPLOYMENT}     c2cc-probe
${PROBE_PORT}           8080


*** Test Cases ***
Probe Namespace Exists
    [Documentation]    Verify the microshift-c2cc namespace exists on all clusters.
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
    [Documentation]    Verify that LastProbeTime is populated after probing starts.
    FOR    ${alias}    IN    cluster-a    cluster-b    cluster-c
        ${stdout}=    Oc On Cluster    ${alias}
        ...    oc get remoteclusters.microshift.io -o jsonpath='{.items[0].status.lastProbeTime}'
        Should Not Be Empty    ${stdout}
    END

RemoteCluster Status Has LastSuccessfulProbe
    [Documentation]    Verify that LastSuccessfulProbe is populated when state is Healthy.
    FOR    ${alias}    IN    cluster-a    cluster-b    cluster-c
        ${stdout}=    Oc On Cluster    ${alias}
        ...    oc get remoteclusters.microshift.io -o jsonpath='{.items[0].status.lastSuccessfulProbe}'
        Should Not Be Empty    ${stdout}
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
    FOR    ${state}    IN    @{states}
        Should Be Equal As Strings    ${state}    ${expected_state}
    END

Compute 11th IP
    [Documentation]    Return the 11th host address in a CIDR (e.g. 10.43.0.0/16 -> 10.43.0.11).
    [Arguments]    ${cidr}
    VAR    ${cmd}=    import ipaddress; n=ipaddress.ip_network('${cidr}', strict=False); print(n[11])
    ${result}=    Process.Run Process    python3    -c    ${cmd}
    Should Be Equal As Integers    ${result.rc}    0
    ${ip}=    Strip String    ${result.stdout}
    RETURN    ${ip}
