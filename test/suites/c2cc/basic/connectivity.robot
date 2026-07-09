*** Settings ***
Documentation       Cross-cluster connectivity tests for C2CC.
...                 Deploys test workloads on all clusters and verifies pod-to-pod
...                 and pod-to-service communication in both directions.

Resource            ../../../resources/microshift-process.resource
Resource            ../../../resources/kubeconfig.resource
Resource            ../../../resources/oc.resource
Resource            ../../../resources/c2cc.resource

Suite Setup         C2CC Suite Setup    deploy_workloads=${TRUE}
Suite Teardown      C2CC Suite Teardown    cleanup_workloads=${TRUE}

Test Tags           c2cc


*** Test Cases ***
Verify Cross Cluster Connectivity
    [Documentation]    Verify pods on all clusters can reach pods/services on all other clusters.
    [Template]    Verify Connectivity Between Clusters
    cluster-a    cluster-b    pod
    cluster-a    cluster-b    service
    cluster-a    cluster-c    pod
    cluster-a    cluster-c    service
    cluster-b    cluster-a    pod
    cluster-b    cluster-a    service
    cluster-b    cluster-c    pod
    cluster-b    cluster-c    service
    cluster-c    cluster-a    pod
    cluster-c    cluster-a    service
    cluster-c    cluster-b    pod
    cluster-c    cluster-b    service

Verify Cross Cluster Source IP Preservation
    [Documentation]    Verify cross cluster traffic preserves source pod IP (no SNAT).
    [Template]    Verify Source IP Preserved Between Clusters
    cluster-a    cluster-b    pod
    cluster-a    cluster-b    service
    cluster-a    cluster-c    pod
    cluster-a    cluster-c    service
    cluster-b    cluster-a    pod
    cluster-b    cluster-a    service
    cluster-b    cluster-c    pod
    cluster-b    cluster-c    service
    cluster-c    cluster-a    pod
    cluster-c    cluster-a    service
    cluster-c    cluster-b    pod
    cluster-c    cluster-b    service

Verify Dual Stack Cross Cluster Connectivity
    [Documentation]    Verify dual-stack pods on all clusters can reach pods/services on all other clusters.
    Skip If    '${CLUSTER_A_POD_CIDR_DUAL}' == ''    Dual-stack CIDRs not configured
    Verify Dual Stack Connectivity Between Clusters    cluster-a    cluster-b    pod
    Verify Dual Stack Connectivity Between Clusters    cluster-a    cluster-b    service
    Verify Dual Stack Connectivity Between Clusters    cluster-a    cluster-c    pod
    Verify Dual Stack Connectivity Between Clusters    cluster-a    cluster-c    service
    Verify Dual Stack Connectivity Between Clusters    cluster-b    cluster-a    pod
    Verify Dual Stack Connectivity Between Clusters    cluster-b    cluster-a    service
    Verify Dual Stack Connectivity Between Clusters    cluster-b    cluster-c    pod
    Verify Dual Stack Connectivity Between Clusters    cluster-b    cluster-c    service
    Verify Dual Stack Connectivity Between Clusters    cluster-c    cluster-a    pod
    Verify Dual Stack Connectivity Between Clusters    cluster-c    cluster-a    service
    Verify Dual Stack Connectivity Between Clusters    cluster-c    cluster-b    pod
    Verify Dual Stack Connectivity Between Clusters    cluster-c    cluster-b    service

Verify Dual Stack Cross Cluster Source IP Preservation
    [Documentation]    Verify dual-stack cross cluster traffic preserves source pod IP (no SNAT).
    Skip If    '${CLUSTER_A_POD_CIDR_DUAL}' == ''    Dual-stack CIDRs not configured
    Verify Dual Stack Source IP Preserved Between Clusters    cluster-a    cluster-b    pod
    Verify Dual Stack Source IP Preserved Between Clusters    cluster-a    cluster-b    service
    Verify Dual Stack Source IP Preserved Between Clusters    cluster-a    cluster-c    pod
    Verify Dual Stack Source IP Preserved Between Clusters    cluster-a    cluster-c    service
    Verify Dual Stack Source IP Preserved Between Clusters    cluster-b    cluster-a    pod
    Verify Dual Stack Source IP Preserved Between Clusters    cluster-b    cluster-a    service
    Verify Dual Stack Source IP Preserved Between Clusters    cluster-b    cluster-c    pod
    Verify Dual Stack Source IP Preserved Between Clusters    cluster-b    cluster-c    service
    Verify Dual Stack Source IP Preserved Between Clusters    cluster-c    cluster-a    pod
    Verify Dual Stack Source IP Preserved Between Clusters    cluster-c    cluster-a    service
    Verify Dual Stack Source IP Preserved Between Clusters    cluster-c    cluster-b    pod
    Verify Dual Stack Source IP Preserved Between Clusters    cluster-c    cluster-b    service
