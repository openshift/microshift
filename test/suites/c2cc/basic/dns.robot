*** Settings ***
Documentation       Cross-cluster DNS tests for C2CC.
...                 Verifies CoreDNS server blocks are injected for remote domains,
...                 DNS resolution works across clusters, and service access via
...                 DNS names works end-to-end.

Resource            ../../../resources/microshift-process.resource
Resource            ../../../resources/kubeconfig.resource
Resource            ../../../resources/oc.resource
Resource            ../../../resources/c2cc.resource

Suite Setup         DNS Suite Setup
Suite Teardown      C2CC Suite Teardown    cleanup_workloads=${TRUE}

Test Tags           c2cc


*** Test Cases ***
Verify Corefile Contains C2CC Server Block
    [Documentation]    Verify every cluster's Corefile has a server block for every other cluster domain.
    [Template]    Verify Corefile Contains C2CC Server Block
    cluster-a    ${CLUSTER_B_DOMAIN}
    cluster-a    ${CLUSTER_C_DOMAIN}
    cluster-b    ${CLUSTER_A_DOMAIN}
    cluster-b    ${CLUSTER_C_DOMAIN}
    cluster-c    ${CLUSTER_A_DOMAIN}
    cluster-c    ${CLUSTER_B_DOMAIN}

Verify Resolve Remote Service DNS
    [Documentation]    Verify pods can resolve a service on all clusters via DNS.
    [Template]    Verify DNS Resolution From Cluster
    cluster-a    hello-microshift.${NAMESPACES}[cluster-b].svc.${DOMAIN_MAP}[cluster-b]
    cluster-a    hello-microshift.${NAMESPACES}[cluster-c].svc.${DOMAIN_MAP}[cluster-c]
    cluster-b    hello-microshift.${NAMESPACES}[cluster-a].svc.${DOMAIN_MAP}[cluster-a]
    cluster-b    hello-microshift.${NAMESPACES}[cluster-c].svc.${DOMAIN_MAP}[cluster-c]
    cluster-c    hello-microshift.${NAMESPACES}[cluster-a].svc.${DOMAIN_MAP}[cluster-a]
    cluster-c    hello-microshift.${NAMESPACES}[cluster-b].svc.${DOMAIN_MAP}[cluster-b]

Verify Remote Service Reachable Via DNS
    [Documentation]    Verify pod on a cluster can reach a service on all clusters using the remote DNS name.
    [Template]    Verify Remote Service Via DNS
    cluster-a    cluster-b
    cluster-a    cluster-c
    cluster-b    cluster-a
    cluster-b    cluster-c
    cluster-c    cluster-a
    cluster-c    cluster-b


*** Keywords ***
DNS Suite Setup
    [Documentation]    Set up clusters and wait for cross-cluster DNS to be healthy
    ...    before running tests. A prior suite may have restarted MicroShift,
    ...    and CoreDNS forward plugin needs time to re-establish upstreams.
    C2CC Suite Setup    deploy_workloads=${TRUE}
    Wait Until Keyword Succeeds    3 minutes    10s
    ...    Verify Full C2CC DNS
