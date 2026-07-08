*** Settings ***
Documentation       Verify C2CC controller sets up all networking infrastructure correctly.
...                 Checks Linux routes, IP rules, nftables bypass, OVN static routes,
...                 and node annotations on all clusters.

Resource            ../../../resources/microshift-process.resource
Resource            ../../../resources/kubeconfig.resource
Resource            ../../../resources/oc.resource
Resource            ../../../resources/c2cc.resource

Suite Setup         C2CC Suite Setup
Suite Teardown      C2CC Suite Teardown

Test Tags           c2cc


*** Test Cases ***
Linux Routes Table 200 Exist
    [Documentation]    Verify routes to remote CIDRs exist in policy routing table 200 between all clusters.
    [Template]    Verify Routes In Table 200
    cluster-a    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}
    cluster-a    ${CLUSTER_C_POD_CIDR}    ${CLUSTER_C_SVC_CIDR}
    cluster-b    ${CLUSTER_A_POD_CIDR}    ${CLUSTER_A_SVC_CIDR}
    cluster-b    ${CLUSTER_C_POD_CIDR}    ${CLUSTER_C_SVC_CIDR}
    cluster-c    ${CLUSTER_A_POD_CIDR}    ${CLUSTER_A_SVC_CIDR}
    cluster-c    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}

IP Rules For Remote CIDRs Exist
    [Documentation]    Verify IP rules at priority 100 direct remote CIDRs to table 200 between all clusters.
    [Template]    Verify IP Rules For Table 200
    cluster-a    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}
    cluster-a    ${CLUSTER_C_POD_CIDR}    ${CLUSTER_C_SVC_CIDR}
    cluster-b    ${CLUSTER_A_POD_CIDR}    ${CLUSTER_A_SVC_CIDR}
    cluster-b    ${CLUSTER_C_POD_CIDR}    ${CLUSTER_C_SVC_CIDR}
    cluster-c    ${CLUSTER_A_POD_CIDR}    ${CLUSTER_A_SVC_CIDR}
    cluster-c    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}

Service Routes Table 201 Exist
    [Documentation]    Verify service routes exist in table 201 on all clusters.
    [Template]    Verify Routes In Table 201
    cluster-a    ${CLUSTER_A_SVC_CIDR}
    cluster-b    ${CLUSTER_B_SVC_CIDR}
    cluster-c    ${CLUSTER_C_SVC_CIDR}

Service IP Rules Exist
    [Documentation]    Verify IP rules at priority 99 for service routing on all clusters.
    [Template]    Verify Service IP Rules
    cluster-a    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}    ${CLUSTER_A_SVC_CIDR}
    cluster-a    ${CLUSTER_C_POD_CIDR}    ${CLUSTER_C_SVC_CIDR}    ${CLUSTER_A_SVC_CIDR}
    cluster-b    ${CLUSTER_A_POD_CIDR}    ${CLUSTER_A_SVC_CIDR}    ${CLUSTER_B_SVC_CIDR}
    cluster-b    ${CLUSTER_C_POD_CIDR}    ${CLUSTER_C_SVC_CIDR}    ${CLUSTER_B_SVC_CIDR}
    cluster-c    ${CLUSTER_A_POD_CIDR}    ${CLUSTER_A_SVC_CIDR}    ${CLUSTER_C_SVC_CIDR}
    cluster-c    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}    ${CLUSTER_C_SVC_CIDR}

NFTables Bypass Rules Exist
    [Documentation]    Verify nftables masquerade bypass rules for remote CIDRs on all clusters.
    [Template]    Verify NFTables Bypass Rules
    cluster-a    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}
    cluster-a    ${CLUSTER_C_POD_CIDR}    ${CLUSTER_C_SVC_CIDR}
    cluster-b    ${CLUSTER_A_POD_CIDR}    ${CLUSTER_A_SVC_CIDR}
    cluster-b    ${CLUSTER_C_POD_CIDR}    ${CLUSTER_C_SVC_CIDR}
    cluster-c    ${CLUSTER_A_POD_CIDR}    ${CLUSTER_A_SVC_CIDR}
    cluster-c    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}

OVN Static Routes Exist
    [Documentation]    Verify OVN NB static routes tagged with microshift-c2cc on all clusters.
    [Template]    Verify OVN Static Routes
    cluster-a    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}
    cluster-a    ${CLUSTER_C_POD_CIDR}    ${CLUSTER_C_SVC_CIDR}
    cluster-b    ${CLUSTER_A_POD_CIDR}    ${CLUSTER_A_SVC_CIDR}
    cluster-b    ${CLUSTER_C_POD_CIDR}    ${CLUSTER_C_SVC_CIDR}
    cluster-c    ${CLUSTER_A_POD_CIDR}    ${CLUSTER_A_SVC_CIDR}
    cluster-c    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}

Node Annotation Set
    [Documentation]    Verify SNAT-exclude annotation contains remote CIDRs on all clusters.
    [Template]    Verify Node SNAT Annotation
    cluster-a    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}
    cluster-a    ${CLUSTER_C_POD_CIDR}    ${CLUSTER_C_SVC_CIDR}
    cluster-b    ${CLUSTER_A_POD_CIDR}    ${CLUSTER_A_SVC_CIDR}
    cluster-b    ${CLUSTER_C_POD_CIDR}    ${CLUSTER_C_SVC_CIDR}
    cluster-c    ${CLUSTER_A_POD_CIDR}    ${CLUSTER_A_SVC_CIDR}
    cluster-c    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}

Dual Stack Linux Routes Table 200 Exist
    [Documentation]    Verify dual-stack routes to remote CIDRs exist in table 200 between all clusters.
    Skip If    '${CLUSTER_A_POD_CIDR_DUAL}' == ''    Dual-stack CIDRs not configured
    ${ip_cmd}=    IP Command For CIDR    ${CLUSTER_A_POD_CIDR_DUAL}
    Verify Routes In Table 200    cluster-a    ${CLUSTER_B_POD_CIDR_DUAL}    ${CLUSTER_B_SVC_CIDR_DUAL}    ${ip_cmd}
    Verify Routes In Table 200    cluster-a    ${CLUSTER_C_POD_CIDR_DUAL}    ${CLUSTER_C_SVC_CIDR_DUAL}    ${ip_cmd}
    Verify Routes In Table 200    cluster-b    ${CLUSTER_A_POD_CIDR_DUAL}    ${CLUSTER_A_SVC_CIDR_DUAL}    ${ip_cmd}
    Verify Routes In Table 200    cluster-b    ${CLUSTER_C_POD_CIDR_DUAL}    ${CLUSTER_C_SVC_CIDR_DUAL}    ${ip_cmd}
    Verify Routes In Table 200    cluster-c    ${CLUSTER_A_POD_CIDR_DUAL}    ${CLUSTER_A_SVC_CIDR_DUAL}    ${ip_cmd}
    Verify Routes In Table 200    cluster-c    ${CLUSTER_B_POD_CIDR_DUAL}    ${CLUSTER_B_SVC_CIDR_DUAL}    ${ip_cmd}

Dual Stack IP Rules For Remote CIDRs Exist
    [Documentation]    Verify dual-stack IP rules direct remote CIDRs to table 200 between all clusters.
    Skip If    '${CLUSTER_A_POD_CIDR_DUAL}' == ''    Dual-stack CIDRs not configured
    ${ip_cmd}=    IP Command For CIDR    ${CLUSTER_A_POD_CIDR_DUAL}
    Verify IP Rules For Table 200    cluster-a    ${CLUSTER_B_POD_CIDR_DUAL}    ${CLUSTER_B_SVC_CIDR_DUAL}    ${ip_cmd}
    Verify IP Rules For Table 200    cluster-a    ${CLUSTER_C_POD_CIDR_DUAL}    ${CLUSTER_C_SVC_CIDR_DUAL}    ${ip_cmd}
    Verify IP Rules For Table 200    cluster-b    ${CLUSTER_A_POD_CIDR_DUAL}    ${CLUSTER_A_SVC_CIDR_DUAL}    ${ip_cmd}
    Verify IP Rules For Table 200    cluster-b    ${CLUSTER_C_POD_CIDR_DUAL}    ${CLUSTER_C_SVC_CIDR_DUAL}    ${ip_cmd}
    Verify IP Rules For Table 200    cluster-c    ${CLUSTER_A_POD_CIDR_DUAL}    ${CLUSTER_A_SVC_CIDR_DUAL}    ${ip_cmd}
    Verify IP Rules For Table 200    cluster-c    ${CLUSTER_B_POD_CIDR_DUAL}    ${CLUSTER_B_SVC_CIDR_DUAL}    ${ip_cmd}

Dual Stack Service Routes Table 201 Exist
    [Documentation]    Verify dual-stack service routes exist in table 201 on all clusters.
    Skip If    '${CLUSTER_A_POD_CIDR_DUAL}' == ''    Dual-stack CIDRs not configured
    ${ip_cmd}=    IP Command For CIDR    ${CLUSTER_A_SVC_CIDR_DUAL}
    Verify Routes In Table 201    cluster-a    ${CLUSTER_A_SVC_CIDR_DUAL}    ${ip_cmd}
    Verify Routes In Table 201    cluster-b    ${CLUSTER_B_SVC_CIDR_DUAL}    ${ip_cmd}
    Verify Routes In Table 201    cluster-c    ${CLUSTER_C_SVC_CIDR_DUAL}    ${ip_cmd}

Dual Stack Service IP Rules Exist
    [Documentation]    Verify dual-stack service IP rules on all clusters.
    Skip If    '${CLUSTER_A_POD_CIDR_DUAL}' == ''    Dual-stack CIDRs not configured
    ${ip_cmd}=    IP Command For CIDR    ${CLUSTER_A_POD_CIDR_DUAL}
    Verify Service IP Rules
    ...    cluster-a
    ...    ${CLUSTER_B_POD_CIDR_DUAL}
    ...    ${CLUSTER_B_SVC_CIDR_DUAL}
    ...    ${CLUSTER_A_SVC_CIDR_DUAL}
    ...    ${ip_cmd}
    Verify Service IP Rules
    ...    cluster-a
    ...    ${CLUSTER_C_POD_CIDR_DUAL}
    ...    ${CLUSTER_C_SVC_CIDR_DUAL}
    ...    ${CLUSTER_A_SVC_CIDR_DUAL}
    ...    ${ip_cmd}
    Verify Service IP Rules
    ...    cluster-b
    ...    ${CLUSTER_A_POD_CIDR_DUAL}
    ...    ${CLUSTER_A_SVC_CIDR_DUAL}
    ...    ${CLUSTER_B_SVC_CIDR_DUAL}
    ...    ${ip_cmd}
    Verify Service IP Rules
    ...    cluster-b
    ...    ${CLUSTER_C_POD_CIDR_DUAL}
    ...    ${CLUSTER_C_SVC_CIDR_DUAL}
    ...    ${CLUSTER_B_SVC_CIDR_DUAL}
    ...    ${ip_cmd}
    Verify Service IP Rules
    ...    cluster-c
    ...    ${CLUSTER_A_POD_CIDR_DUAL}
    ...    ${CLUSTER_A_SVC_CIDR_DUAL}
    ...    ${CLUSTER_C_SVC_CIDR_DUAL}
    ...    ${ip_cmd}
    Verify Service IP Rules
    ...    cluster-c
    ...    ${CLUSTER_B_POD_CIDR_DUAL}
    ...    ${CLUSTER_B_SVC_CIDR_DUAL}
    ...    ${CLUSTER_C_SVC_CIDR_DUAL}
    ...    ${ip_cmd}

Dual Stack NFTables Bypass Rules Exist
    [Documentation]    Verify dual-stack nftables masquerade bypass rules on all clusters.
    Skip If    '${CLUSTER_A_POD_CIDR_DUAL}' == ''    Dual-stack CIDRs not configured
    Verify NFTables Bypass Rules    cluster-a    ${CLUSTER_B_POD_CIDR_DUAL}    ${CLUSTER_B_SVC_CIDR_DUAL}
    Verify NFTables Bypass Rules    cluster-a    ${CLUSTER_C_POD_CIDR_DUAL}    ${CLUSTER_C_SVC_CIDR_DUAL}
    Verify NFTables Bypass Rules    cluster-b    ${CLUSTER_A_POD_CIDR_DUAL}    ${CLUSTER_A_SVC_CIDR_DUAL}
    Verify NFTables Bypass Rules    cluster-b    ${CLUSTER_C_POD_CIDR_DUAL}    ${CLUSTER_C_SVC_CIDR_DUAL}
    Verify NFTables Bypass Rules    cluster-c    ${CLUSTER_A_POD_CIDR_DUAL}    ${CLUSTER_A_SVC_CIDR_DUAL}
    Verify NFTables Bypass Rules    cluster-c    ${CLUSTER_B_POD_CIDR_DUAL}    ${CLUSTER_B_SVC_CIDR_DUAL}

Dual Stack OVN Static Routes Exist
    [Documentation]    Verify dual-stack OVN NB static routes tagged with microshift-c2cc on all clusters.
    Skip If    '${CLUSTER_A_POD_CIDR_DUAL}' == ''    Dual-stack CIDRs not configured
    Verify OVN Static Routes    cluster-a    ${CLUSTER_B_POD_CIDR_DUAL}    ${CLUSTER_B_SVC_CIDR_DUAL}
    Verify OVN Static Routes    cluster-a    ${CLUSTER_C_POD_CIDR_DUAL}    ${CLUSTER_C_SVC_CIDR_DUAL}
    Verify OVN Static Routes    cluster-b    ${CLUSTER_A_POD_CIDR_DUAL}    ${CLUSTER_A_SVC_CIDR_DUAL}
    Verify OVN Static Routes    cluster-b    ${CLUSTER_C_POD_CIDR_DUAL}    ${CLUSTER_C_SVC_CIDR_DUAL}
    Verify OVN Static Routes    cluster-c    ${CLUSTER_A_POD_CIDR_DUAL}    ${CLUSTER_A_SVC_CIDR_DUAL}
    Verify OVN Static Routes    cluster-c    ${CLUSTER_B_POD_CIDR_DUAL}    ${CLUSTER_B_SVC_CIDR_DUAL}

Dual Stack Node Annotation Set
    [Documentation]    Verify SNAT-exclude annotation contains dual-stack remote CIDRs on all clusters.
    Skip If    '${CLUSTER_A_POD_CIDR_DUAL}' == ''    Dual-stack CIDRs not configured
    Verify Node SNAT Annotation    cluster-a    ${CLUSTER_B_POD_CIDR_DUAL}    ${CLUSTER_B_SVC_CIDR_DUAL}
    Verify Node SNAT Annotation    cluster-a    ${CLUSTER_C_POD_CIDR_DUAL}    ${CLUSTER_C_SVC_CIDR_DUAL}
    Verify Node SNAT Annotation    cluster-b    ${CLUSTER_A_POD_CIDR_DUAL}    ${CLUSTER_A_SVC_CIDR_DUAL}
    Verify Node SNAT Annotation    cluster-b    ${CLUSTER_C_POD_CIDR_DUAL}    ${CLUSTER_C_SVC_CIDR_DUAL}
    Verify Node SNAT Annotation    cluster-c    ${CLUSTER_A_POD_CIDR_DUAL}    ${CLUSTER_A_SVC_CIDR_DUAL}
    Verify Node SNAT Annotation    cluster-c    ${CLUSTER_B_POD_CIDR_DUAL}    ${CLUSTER_B_SVC_CIDR_DUAL}
