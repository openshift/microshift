*** Settings ***
Documentation       Cross-cluster DNS tests for C2CC.
...                 Verifies CoreDNS server blocks are injected for remote domains,
...                 DNS resolution works across clusters, and service access via
...                 DNS names works end-to-end.

Resource            ../../resources/microshift-process.resource
Resource            ../../resources/kubeconfig.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/c2cc.resource

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           c2cc


*** Test Cases ***
Test Corefile Contains C2CC Server Block
    [Documentation]    Verify every cluster's Corefile has a server block for every other cluster domain.
    [Template]    Verify Corefile Contains C2CC Server Block
    cluster-a    ${CLUSTER_B_DOMAIN}
    cluster-a    ${CLUSTER_C_DOMAIN}
    cluster-b    ${CLUSTER_A_DOMAIN}
    cluster-b    ${CLUSTER_C_DOMAIN}
    cluster-c    ${CLUSTER_A_DOMAIN}
    cluster-c    ${CLUSTER_B_DOMAIN}

Test Resolve Remote Service DNS
    [Documentation]    Verify pods can resolve a service on all clusters via DNS.
    [Template]    DNS Resolve From Cluster
    cluster-a    hello-microshift.${NAMESPACES}[cluster-b].svc.${DOMAIN_MAP}[cluster-b]
    cluster-a    hello-microshift.${NAMESPACES}[cluster-c].svc.${DOMAIN_MAP}[cluster-c]
    cluster-b    hello-microshift.${NAMESPACES}[cluster-a].svc.${DOMAIN_MAP}[cluster-a]
    cluster-b    hello-microshift.${NAMESPACES}[cluster-c].svc.${DOMAIN_MAP}[cluster-c]
    cluster-c    hello-microshift.${NAMESPACES}[cluster-a].svc.${DOMAIN_MAP}[cluster-a]
    cluster-c    hello-microshift.${NAMESPACES}[cluster-b].svc.${DOMAIN_MAP}[cluster-b]

Test Curl Remote Service Via DNS
    [Documentation]    Verify pod on a cluster can reach a service on all clusters using the remote DNS name.
    [Template]    Curl Remote Service Via DNS
    cluster-a    cluster-b
    cluster-a    cluster-c
    cluster-b    cluster-a
    cluster-b    cluster-c
    cluster-c    cluster-a
    cluster-c    cluster-b


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

Teardown
    [Documentation]    Remove test workloads and close connections.
    Cleanup Test Workloads
    Teardown All Remote Clusters
    Remove Kubeconfig
    Logout MicroShift Host
