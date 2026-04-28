*** Settings ***
Documentation       Verify C2CC controller sets up all networking infrastructure correctly.
...                 Checks Linux routes, IP rules, nftables bypass, OVN static routes,
...                 node annotations, and network policies on both clusters.

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
Linux Routes Table 200 Exist On Cluster A
    [Documentation]    Verify routes to remote CIDRs exist in policy routing table 200 on Cluster A.
    Verify Routes In Table 200    cluster-a    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}

Linux Routes Table 200 Exist On Cluster B
    [Documentation]    Verify routes to remote CIDRs exist in policy routing table 200 on Cluster B.
    Verify Routes In Table 200    cluster-b    ${CLUSTER_A_POD_CIDR}    ${CLUSTER_A_SVC_CIDR}

IP Rules For Remote CIDRs Exist On Cluster A
    [Documentation]    Verify IP rules at priority 100 direct remote CIDRs to table 200 on Cluster A.
    Verify IP Rules For Table 200    cluster-a    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}

IP Rules For Remote CIDRs Exist On Cluster B
    [Documentation]    Verify IP rules at priority 100 direct remote CIDRs to table 200 on Cluster B.
    Verify IP Rules For Table 200    cluster-b    ${CLUSTER_A_POD_CIDR}    ${CLUSTER_A_SVC_CIDR}

Service Routes Table 201 Exist On Cluster A
    [Documentation]    Verify service routes exist in table 201 on Cluster A.
    Verify Routes In Table 201    cluster-a    ${CLUSTER_A_SVC_CIDR}

Service Routes Table 201 Exist On Cluster B
    [Documentation]    Verify service routes exist in table 201 on Cluster B.
    Verify Routes In Table 201    cluster-b    ${CLUSTER_B_SVC_CIDR}

Service IP Rules Exist On Cluster A
    [Documentation]    Verify IP rules at priority 99 for service routing on Cluster A.
    Verify Service IP Rules    cluster-a    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}    ${CLUSTER_A_SVC_CIDR}

Service IP Rules Exist On Cluster B
    [Documentation]    Verify IP rules at priority 99 for service routing on Cluster B.
    Verify Service IP Rules    cluster-b    ${CLUSTER_A_POD_CIDR}    ${CLUSTER_A_SVC_CIDR}    ${CLUSTER_B_SVC_CIDR}

NFTables Bypass Rules Exist On Cluster A
    [Documentation]    Verify nftables masquerade bypass rules for remote CIDRs on Cluster A.
    Verify NFTables Bypass Rules    cluster-a    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}

NFTables Bypass Rules Exist On Cluster B
    [Documentation]    Verify nftables masquerade bypass rules for remote CIDRs on Cluster B.
    Verify NFTables Bypass Rules    cluster-b    ${CLUSTER_A_POD_CIDR}    ${CLUSTER_A_SVC_CIDR}

OVN Static Routes Exist On Cluster A
    [Documentation]    Verify OVN NB static routes tagged with microshift-c2cc on Cluster A.
    Verify OVN Static Routes    cluster-a    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}

OVN Static Routes Exist On Cluster B
    [Documentation]    Verify OVN NB static routes tagged with microshift-c2cc on Cluster B.
    Verify OVN Static Routes    cluster-b    ${CLUSTER_A_POD_CIDR}    ${CLUSTER_A_SVC_CIDR}

Node Annotation Set On Cluster A
    [Documentation]    Verify SNAT-exclude annotation contains remote CIDRs on Cluster A.
    Verify Node SNAT Annotation    cluster-a    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}

Node Annotation Set On Cluster B
    [Documentation]    Verify SNAT-exclude annotation contains remote CIDRs on Cluster B.
    Verify Node SNAT Annotation    cluster-b    ${CLUSTER_A_POD_CIDR}    ${CLUSTER_A_SVC_CIDR}

Network Policy Exists On Cluster A
    [Documentation]    Verify C2CC network policy exists in default namespace on Cluster A.
    Verify C2CC Network Policy    cluster-a

Network Policy Exists On Cluster B
    [Documentation]    Verify C2CC network policy exists in default namespace on Cluster B.
    Verify C2CC Network Policy    cluster-b


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
