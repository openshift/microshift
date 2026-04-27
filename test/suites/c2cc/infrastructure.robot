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

Verify Routes In Table 200
    [Documentation]    Check that routes for the given CIDRs exist in table 200.
    [Arguments]    ${alias}    ${remote_pod_cidr}    ${remote_svc_cidr}
    ${stdout}=    Command On Cluster    ${alias}    ip route show table 200
    Should Contain    ${stdout}    ${remote_pod_cidr}
    Should Contain    ${stdout}    ${remote_svc_cidr}

Verify IP Rules For Table 200
    [Documentation]    Check that IP rules at priority 100 exist for the given CIDRs.
    [Arguments]    ${alias}    ${remote_pod_cidr}    ${remote_svc_cidr}
    ${stdout}=    Command On Cluster    ${alias}    ip rule show
    Should Contain    ${stdout}    to ${remote_pod_cidr} lookup 200
    Should Contain    ${stdout}    to ${remote_svc_cidr} lookup 200

Verify Routes In Table 201
    [Documentation]    Check that service routes exist in table 201 for the local service CIDR.
    [Arguments]    ${alias}    ${local_svc_cidr}
    ${stdout}=    Command On Cluster    ${alias}    ip route show table 201
    Should Contain    ${stdout}    ${local_svc_cidr}

Verify Service IP Rules
    [Documentation]    Check that IP rules at priority 99 exist for cross-cluster service routing.
    [Arguments]    ${alias}    ${remote_pod_cidr}    ${remote_svc_cidr}    ${local_svc_cidr}
    ${stdout}=    Command On Cluster    ${alias}    ip rule show
    Should Contain    ${stdout}    from ${remote_pod_cidr} to ${local_svc_cidr} lookup 201
    Should Contain    ${stdout}    from ${remote_svc_cidr} to ${local_svc_cidr} lookup 201

Verify NFTables Bypass Rules
    [Documentation]    Check that nftables masquerade bypass rules exist for remote CIDRs.
    [Arguments]    ${alias}    ${remote_pod_cidr}    ${remote_svc_cidr}
    ${stdout}=    Command On Cluster    ${alias}
    ...    nft list chain inet ovn-kubernetes ovn-kube-pod-subnet-masq
    Should Contain    ${stdout}    c2cc-no-masq:${remote_pod_cidr}
    Should Contain    ${stdout}    c2cc-no-masq:${remote_svc_cidr}

Verify OVN Static Routes
    [Documentation]    Check that OVN NB static routes tagged with microshift-c2cc exist for remote CIDRs.
    [Arguments]    ${alias}    ${remote_pod_cidr}    ${remote_svc_cidr}
    ${pod}=    Oc On Cluster    ${alias}
    ...    oc get pod -n openshift-ovn-kubernetes -l app=ovnkube-master -o jsonpath='{.items[0].metadata.name}'
    ${stdout}=    Oc On Cluster    ${alias}
    ...    oc exec -n openshift-ovn-kubernetes ${pod} -- ovn-nbctl find Logical_Router_Static_Route external_ids:k8s.ovn.org/owner-controller=microshift-c2cc
    Should Contain    ${stdout}    ${remote_pod_cidr}
    Should Contain    ${stdout}    ${remote_svc_cidr}

Verify Node SNAT Annotation
    [Documentation]    Check that the node SNAT-exclude annotation contains the remote CIDRs.
    [Arguments]    ${alias}    ${remote_pod_cidr}    ${remote_svc_cidr}
    ${stdout}=    Oc On Cluster    ${alias}
    ...    oc get node -o json | python3 -c "import json,sys; print(json.load(sys.stdin)['items'][0]['metadata']['annotations'].get('k8s.ovn.org/node-ingress-snat-exclude-subnets',''))"
    Should Contain    ${stdout}    ${remote_pod_cidr}
    Should Contain    ${stdout}    ${remote_svc_cidr}

Verify C2CC Network Policy
    [Documentation]    Check that the C2CC network policy exists in the default namespace.
    [Arguments]    ${alias}
    ${stdout}=    Oc On Cluster    ${alias}
    ...    oc get networkpolicy c2cc-allow-remote-pods -n default -o jsonpath='{.metadata.labels.app\\.kubernetes\\.io/managed-by}'
    Should Be Equal As Strings    ${stdout}    microshift-c2cc
