*** Settings ***
Documentation       Verify C2CC survives VM reboots in escalating scenarios.
...                 Tests single cluster, two clusters simultaneously, and all three
...                 clusters simultaneously. After each reboot cycle, full-stack
...                 verification confirms connectivity, infrastructure, health probes,
...                 and DNS all recover.

Resource            ../../resources/microshift-process.resource
Resource            ../../resources/kubeconfig.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/c2cc.resource
Resource            ../../resources/ostree-health.resource

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           c2cc


*** Test Cases ***
Reboot Single Cluster
    [Documentation]    Reboot cluster-a while cluster-b and cluster-c remain up.
    ...    Verifies that the rebooted cluster re-establishes C2CC connectivity
    ...    with both peers.
    [Setup]    Ensure All Clusters Healthy
    Reboot Clusters Simultaneously    cluster-a
    Wait For Clusters Ready    cluster-a
    Verify Full C2CC Stack

Reboot Two Clusters Simultaneously
    [Documentation]    Reboot cluster-b and cluster-c at the same time.
    ...    The surviving cluster-a must wait for both peers to recover.
    ...    The two rebooted clusters must also reconnect with each other.
    [Setup]    Ensure All Clusters Healthy
    Reboot Clusters Simultaneously    cluster-b    cluster-c
    Wait For Clusters Ready    cluster-b    cluster-c
    Verify Full C2CC Stack

Reboot All Three Clusters Simultaneously
    [Documentation]    Reboot all three clusters at once.
    ...    Every cluster starts from scratch simultaneously — no running peer
    ...    to reference. All must independently reconstruct C2CC state.
    [Setup]    Ensure All Clusters Healthy
    Reboot Clusters Simultaneously    cluster-a    cluster-b    cluster-c
    Wait For Clusters Ready    cluster-a    cluster-b    cluster-c
    Verify Full C2CC Stack


*** Keywords ***
Setup
    [Documentation]    Set up clusters and deploy test workloads on all.
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig
    Register Local Cluster    cluster-a
    Register Remote Cluster    cluster-b    ${HOST2_IP}    ${HOST2_SSH_PORT}    ${KUBECONFIG_B}
    Register Remote Cluster    cluster-c    ${HOST3_IP}    ${HOST3_SSH_PORT}    ${KUBECONFIG_C}
    Deploy Test Workloads
    Verify Full C2CC Stack

Teardown
    [Documentation]    Remove test workloads and close connections.
    Cleanup Test Workloads
    Teardown All Remote Clusters
    Remove Kubeconfig
    Logout MicroShift Host

Wait For Clusters Ready
    [Documentation]    Wait for rebooted clusters to finish greenboot and for test pods
    ...    and service endpoints to become ready.
    [Arguments]    @{cluster_aliases}
    Wait Until Keyword Succeeds    10m    15s
    ...    All Clusters Greenboot Exited    @{cluster_aliases}
    Wait For Test Pods
    Wait For Service Endpoints

All Clusters Greenboot Exited
    [Documentation]    Check that greenboot has exited on all given clusters.
    ...    Fails if any cluster has not finished yet, causing the caller's
    ...    Wait Until Keyword Succeeds to retry.
    [Arguments]    @{cluster_aliases}
    FOR    ${alias}    IN    @{cluster_aliases}
        ${conn_id}=    Get From Dictionary    ${C2CC_SSH_IDS}    ${alias}
        SSHLibrary.Switch Connection    ${conn_id}
        Greenboot Health Check Exited
    END

Verify Full C2CC Stack
    [Documentation]    Comprehensive verification of all C2CC components across all clusters.
    Wait Until Keyword Succeeds    10m    10s    Verify C2CC Connectivity
    Wait Until Keyword Succeeds    10m    10s    Verify C2CC Infrastructure
    Wait Until Keyword Succeeds    10m    10s    Verify C2CC Health Probes
    Wait Until Keyword Succeeds    10m    10s    Verify C2CC DNS

Verify C2CC Connectivity
    [Documentation]    Verify pod-to-pod, pod-to-service connectivity and source IP preservation
    ...    across all 6 cluster pairs.
    FOR    ${src}    ${dst}    IN
    ...    cluster-a    cluster-b
    ...    cluster-a    cluster-c
    ...    cluster-b    cluster-a
    ...    cluster-b    cluster-c
    ...    cluster-c    cluster-a
    ...    cluster-c    cluster-b
        Test Connectivity Between Clusters    ${src}    ${dst}    pod
        Test Connectivity Between Clusters    ${src}    ${dst}    service
        Test Source IP Preserved Between Clusters    ${src}    ${dst}    pod
        Test Source IP Preserved Between Clusters    ${src}    ${dst}    service
    END

Verify C2CC Infrastructure
    [Documentation]    Verify routes, IP rules, nftables, OVN static routes,
    ...    and node annotations for all cluster-peer combinations.
    Verify Infra For Remote Peer    cluster-a    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}    ${CLUSTER_A_SVC_CIDR}
    Verify Infra For Remote Peer    cluster-a    ${CLUSTER_C_POD_CIDR}    ${CLUSTER_C_SVC_CIDR}    ${CLUSTER_A_SVC_CIDR}
    Verify Infra For Remote Peer    cluster-b    ${CLUSTER_A_POD_CIDR}    ${CLUSTER_A_SVC_CIDR}    ${CLUSTER_B_SVC_CIDR}
    Verify Infra For Remote Peer    cluster-b    ${CLUSTER_C_POD_CIDR}    ${CLUSTER_C_SVC_CIDR}    ${CLUSTER_B_SVC_CIDR}
    Verify Infra For Remote Peer    cluster-c    ${CLUSTER_A_POD_CIDR}    ${CLUSTER_A_SVC_CIDR}    ${CLUSTER_C_SVC_CIDR}
    Verify Infra For Remote Peer    cluster-c    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}    ${CLUSTER_C_SVC_CIDR}

Verify Infra For Remote Peer
    [Documentation]    Verify all infrastructure components on a cluster for one remote peer.
    [Arguments]    ${alias}    ${remote_pod_cidr}    ${remote_svc_cidr}    ${local_svc_cidr}
    Verify Routes In Table 200    ${alias}    ${remote_pod_cidr}    ${remote_svc_cidr}
    Verify IP Rules For Table 200    ${alias}    ${remote_pod_cidr}    ${remote_svc_cidr}
    Verify Routes In Table 201    ${alias}    ${local_svc_cidr}
    Verify Service IP Rules    ${alias}    ${remote_pod_cidr}    ${remote_svc_cidr}    ${local_svc_cidr}
    Verify NFTables Bypass Rules    ${alias}    ${remote_pod_cidr}    ${remote_svc_cidr}
    Verify OVN Static Routes    ${alias}    ${remote_pod_cidr}    ${remote_svc_cidr}
    Verify Node SNAT Annotation    ${alias}    ${remote_pod_cidr}    ${remote_svc_cidr}
    Verify C2CC Tracking Annotation    ${alias}    ${remote_pod_cidr}    ${remote_svc_cidr}

Verify C2CC Health Probes
    [Documentation]    Verify all RemoteCluster CRs report Healthy with populated timestamps.
    FOR    ${alias}    IN    cluster-a    cluster-b    cluster-c
        Verify RemoteCluster State    ${alias}    Healthy
        ${stdout}=    Oc On Cluster    ${alias}
        ...    oc get remoteclusters.microshift.io -o jsonpath='{.items[*].status.lastProbeTime}'
        Should Not Be Empty    ${stdout}
        ${stdout}=    Oc On Cluster    ${alias}
        ...    oc get remoteclusters.microshift.io -o jsonpath='{.items[*].status.lastSuccessfulProbe}'
        Should Not Be Empty    ${stdout}
    END

Verify C2CC DNS
    [Documentation]    Verify CoreDNS Corefile contains C2CC server blocks and
    ...    cross-cluster DNS resolution works for all pairs.
    Verify Corefile Contains C2CC Server Block    cluster-a    ${CLUSTER_B_DOMAIN}
    Verify Corefile Contains C2CC Server Block    cluster-a    ${CLUSTER_C_DOMAIN}
    Verify Corefile Contains C2CC Server Block    cluster-b    ${CLUSTER_A_DOMAIN}
    Verify Corefile Contains C2CC Server Block    cluster-b    ${CLUSTER_C_DOMAIN}
    Verify Corefile Contains C2CC Server Block    cluster-c    ${CLUSTER_A_DOMAIN}
    Verify Corefile Contains C2CC Server Block    cluster-c    ${CLUSTER_B_DOMAIN}
    Curl Remote Service Via DNS    cluster-a    cluster-b
    Curl Remote Service Via DNS    cluster-a    cluster-c
    Curl Remote Service Via DNS    cluster-b    cluster-a
    Curl Remote Service Via DNS    cluster-b    cluster-c
    Curl Remote Service Via DNS    cluster-c    cluster-a
    Curl Remote Service Via DNS    cluster-c    cluster-b

Ensure All Clusters Healthy
    [Documentation]    Pre-condition: all clusters must have Healthy RemoteCluster CRs.
    Verify All RemoteClusters Healthy
