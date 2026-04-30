*** Settings ***
Documentation       Negative/fault-injection tests for C2CC controller reconciliation.
...                 Each test deletes or corrupts a specific piece of C2CC networking state,
...                 then waits for the controller to detect the disruption and self-heal.

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-host.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/kubeconfig.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/c2cc.resource

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           c2cc


*** Variables ***
${RECONCILE_TIMEOUT}    30s
${RECONCILE_RETRY}      5s


*** Test Cases ***
Reconcile Linux Route In Table 200 After Deletion
    [Documentation]    Delete a route from table 200, verify the controller restores it.
    Delete Route From Table 200 On Cluster    cluster-a    ${CLUSTER_B_POD_CIDR}
    Wait Until Keyword Succeeds    ${RECONCILE_TIMEOUT}    ${RECONCILE_RETRY}
    ...    Verify Routes In Table 200    cluster-a    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}

Reconcile IP Rule For Table 200 After Deletion
    [Documentation]    Delete an IP rule for table 200, verify the controller restores it.
    Delete IP Rule For Table 200 On Cluster    cluster-a    ${CLUSTER_B_POD_CIDR}
    Wait Until Keyword Succeeds    ${RECONCILE_TIMEOUT}    ${RECONCILE_RETRY}
    ...    Verify IP Rules For Table 200    cluster-a    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}

Reconcile Service Route In Table 201 After Deletion
    [Documentation]    Delete a service route from table 201, verify the controller restores it.
    Delete Service Route From Table 201 On Cluster    cluster-a    ${CLUSTER_A_SVC_CIDR}
    Wait Until Keyword Succeeds    ${RECONCILE_TIMEOUT}    ${RECONCILE_RETRY}
    ...    Verify Routes In Table 201    cluster-a    ${CLUSTER_A_SVC_CIDR}

Reconcile Service IP Rule After Deletion
    [Documentation]    Delete a service IP rule, verify the controller restores it.
    Delete Service IP Rule On Cluster    cluster-a    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_A_SVC_CIDR}
    Wait Until Keyword Succeeds    ${RECONCILE_TIMEOUT}    ${RECONCILE_RETRY}
    ...    Verify Service IP Rules    cluster-a    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}    ${CLUSTER_A_SVC_CIDR}

Reconcile NFTables Bypass Rule After Deletion
    [Documentation]    Delete an nftables bypass rule, verify the controller restores it.
    Delete NFTables C2CC Rule On Cluster    cluster-a    ${CLUSTER_B_POD_CIDR}
    Wait Until Keyword Succeeds    ${RECONCILE_TIMEOUT}    ${RECONCILE_RETRY}
    ...    Verify NFTables Bypass Rules    cluster-a    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}

Reconcile OVN Static Route After Deletion
    [Documentation]    Delete an OVN static route, verify the controller restores it.
    Delete OVN C2CC Route On Cluster    cluster-a    ${CLUSTER_B_POD_CIDR}
    Wait Until Keyword Succeeds    ${RECONCILE_TIMEOUT}    ${RECONCILE_RETRY}
    ...    Verify OVN Static Routes    cluster-a    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}

Reconcile Node SNAT Annotation After Corruption
    [Documentation]    Overwrite the SNAT annotation with empty, verify the controller restores it.
    Corrupt Node SNAT Annotation On Cluster    cluster-a
    Wait Until Keyword Succeeds    ${RECONCILE_TIMEOUT}    ${RECONCILE_RETRY}
    ...    Verify Node SNAT Annotation    cluster-a    ${CLUSTER_B_POD_CIDR}    ${CLUSTER_B_SVC_CIDR}

Reconcile NetworkPolicy After Deletion
    [Documentation]    Delete the C2CC NetworkPolicy, verify the controller re-creates it.
    Delete C2CC NetworkPolicy On Cluster    cluster-a
    Wait Until Keyword Succeeds    ${RECONCILE_TIMEOUT}    ${RECONCILE_RETRY}
    ...    Verify C2CC Network Policy    cluster-a


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

Get Node Name On Cluster
    [Documentation]    Get the name of the first node on the given cluster.
    [Arguments]    ${alias}
    ${node}=    Oc On Cluster    ${alias}
    ...    oc get nodes -o jsonpath='{.items[0].metadata.name}'
    RETURN    ${node}

Delete Route From Table 200 On Cluster
    [Documentation]    Delete a specific route from policy routing table 200.
    [Arguments]    ${alias}    ${cidr}
    Disruptive Command On Cluster    ${alias}    ip route del ${cidr} table 200

Delete IP Rule For Table 200 On Cluster
    [Documentation]    Delete an IP rule directing traffic to table 200.
    [Arguments]    ${alias}    ${cidr}
    Disruptive Command On Cluster    ${alias}    ip rule del to ${cidr} lookup 200

Delete Service Route From Table 201 On Cluster
    [Documentation]    Delete a service route from table 201.
    [Arguments]    ${alias}    ${cidr}
    Disruptive Command On Cluster    ${alias}    ip route del ${cidr} table 201

Delete Service IP Rule On Cluster
    [Documentation]    Delete a service IP rule from table 201.
    [Arguments]    ${alias}    ${from_cidr}    ${to_cidr}
    Disruptive Command On Cluster    ${alias}
    ...    ip rule del from ${from_cidr} to ${to_cidr} lookup 201

Delete NFTables C2CC Rule On Cluster
    [Documentation]    Delete an nftables bypass rule by discovering its handle.
    [Arguments]    ${alias}    ${cidr}
    ${handle}=    Command On Cluster    ${alias}
    ...    nft list chain inet ovn-kubernetes ovn-kube-pod-subnet-masq -a | grep 'c2cc-no-masq:${cidr}' | awk '/# handle/{print $NF}'
    Disruptive Command On Cluster    ${alias}
    ...    nft delete rule inet ovn-kubernetes ovn-kube-pod-subnet-masq handle ${handle}

Delete OVN C2CC Route On Cluster
    [Documentation]    Delete an OVN static route for a CIDR from the gateway router.
    [Arguments]    ${alias}    ${cidr}
    ${pod}=    Oc On Cluster    ${alias}
    ...    oc get pod -n openshift-ovn-kubernetes -l app=ovnkube-master -o jsonpath='{.items[0].metadata.name}'
    ${node}=    Get Node Name On Cluster    ${alias}
    Oc On Cluster    ${alias}
    ...    oc exec -n openshift-ovn-kubernetes ${pod} -- ovn-nbctl lr-route-del GR_${node} ${cidr}

Corrupt Node SNAT Annotation On Cluster
    [Documentation]    Overwrite the SNAT-exclude annotation with an empty list.
    [Arguments]    ${alias}
    ${node}=    Get Node Name On Cluster    ${alias}
    Oc On Cluster    ${alias}
    ...    oc annotate node ${node} k8s.ovn.org/node-ingress-snat-exclude-subnets='[]' --overwrite

Delete C2CC NetworkPolicy On Cluster
    [Documentation]    Delete the C2CC NetworkPolicy from the default namespace.
    [Arguments]    ${alias}
    Oc On Cluster    ${alias}
    ...    oc delete networkpolicy c2cc-allow-remote-pods -n default
