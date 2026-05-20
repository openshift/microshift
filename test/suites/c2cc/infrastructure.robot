*** Settings ***
Documentation       Verify C2CC controller sets up all networking infrastructure correctly.
...                 Checks Linux routes, IP rules, nftables bypass, OVN static routes,
...                 and node annotations on both clusters.

Resource            ../../resources/microshift-process.resource
Resource            ../../resources/kubeconfig.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/c2cc.resource

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           c2cc


*** Test Cases ***
Linux Routes Table 200 Exist
    [Template]    Verify Routes In Table 200
    cluster-a    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}
    cluster-a    ${CLUSTER_C_POD_CIDR}    ${CLUSTER_C_SVC_CIDR}
    cluster-b    ${CLUSTER_A_POD_CIDR}    ${CLUSTER_A_SVC_CIDR}
    cluster-b    ${CLUSTER_C_POD_CIDR}    ${CLUSTER_C_SVC_CIDR}
    cluster-c    ${CLUSTER_A_POD_CIDR}    ${CLUSTER_A_SVC_CIDR}
    cluster-c    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}

IP Rules For Remote CIDRs Exist
    [Template]    Verify IP Rules For Table 200
    cluster-a    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}
    cluster-a    ${CLUSTER_C_POD_CIDR}    ${CLUSTER_C_SVC_CIDR}
    cluster-b    ${CLUSTER_A_POD_CIDR}    ${CLUSTER_A_SVC_CIDR}
    cluster-b    ${CLUSTER_C_POD_CIDR}    ${CLUSTER_C_SVC_CIDR}
    cluster-c    ${CLUSTER_A_POD_CIDR}    ${CLUSTER_A_SVC_CIDR}
    cluster-c    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}

Service Routes Table 201 Exist
    [Template]    Verify Routes In Table 201
    cluster-a    ${CLUSTER_A_SVC_CIDR}
    cluster-b    ${CLUSTER_B_SVC_CIDR}
    cluster-c    ${CLUSTER_C_SVC_CIDR}

Service IP Rules Exist
    [Template]    Verify Service IP Rules
    cluster-a    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}    ${CLUSTER_A_SVC_CIDR}
    cluster-a    ${CLUSTER_C_POD_CIDR}    ${CLUSTER_C_SVC_CIDR}    ${CLUSTER_A_SVC_CIDR}
    cluster-b    ${CLUSTER_A_POD_CIDR}    ${CLUSTER_A_SVC_CIDR}    ${CLUSTER_B_SVC_CIDR}
    cluster-b    ${CLUSTER_C_POD_CIDR}    ${CLUSTER_C_SVC_CIDR}    ${CLUSTER_B_SVC_CIDR}
    cluster-c    ${CLUSTER_A_POD_CIDR}    ${CLUSTER_A_SVC_CIDR}    ${CLUSTER_C_SVC_CIDR}
    cluster-c    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}    ${CLUSTER_C_SVC_CIDR}

NFTables Bypass Rules Exist
    [Template]    Verify NFTables Bypass Rules
    cluster-a    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}
    cluster-a    ${CLUSTER_C_POD_CIDR}    ${CLUSTER_C_SVC_CIDR}
    cluster-b    ${CLUSTER_A_POD_CIDR}    ${CLUSTER_A_SVC_CIDR}
    cluster-b    ${CLUSTER_C_POD_CIDR}    ${CLUSTER_C_SVC_CIDR}
    cluster-c    ${CLUSTER_A_POD_CIDR}    ${CLUSTER_A_SVC_CIDR}
    cluster-c    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}

OVN Static Routes Exist
    [Template]    Verify OVN Static Routes
    cluster-a    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}
    cluster-a    ${CLUSTER_C_POD_CIDR}    ${CLUSTER_C_SVC_CIDR}
    cluster-b    ${CLUSTER_A_POD_CIDR}    ${CLUSTER_A_SVC_CIDR}
    cluster-b    ${CLUSTER_C_POD_CIDR}    ${CLUSTER_C_SVC_CIDR}
    cluster-c    ${CLUSTER_A_POD_CIDR}    ${CLUSTER_A_SVC_CIDR}
    cluster-c    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}

Node Annotation Set
    [Template]    Verify Node SNAT Annotation
    cluster-a    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}
    cluster-a    ${CLUSTER_C_POD_CIDR}    ${CLUSTER_C_SVC_CIDR}
    cluster-b    ${CLUSTER_A_POD_CIDR}    ${CLUSTER_A_SVC_CIDR}
    cluster-b    ${CLUSTER_C_POD_CIDR}    ${CLUSTER_C_SVC_CIDR}
    cluster-c    ${CLUSTER_A_POD_CIDR}    ${CLUSTER_A_SVC_CIDR}
    cluster-c    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}


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
