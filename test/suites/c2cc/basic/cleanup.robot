*** Settings ***
Documentation       Verify C2CC cleanup when the feature is disabled.
...                 Removes the C2CC config drop-in on Cluster A, restarts MicroShift,
...                 and verifies all C2CC networking state has been cleaned up.

Resource            ../../../resources/microshift-process.resource
Resource            ../../../resources/kubeconfig.resource
Resource            ../../../resources/oc.resource
Resource            ../../../resources/c2cc.resource

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           c2cc


*** Variables ***
${CLEANUP_TIMEOUT}      360s
${CLEANUP_RETRY}        30s
${C2CC_CONFIG_PATH}     /etc/microshift/config.d/50-c2cc.yaml


*** Test Cases ***
No Linux Routes In Table 200 After Disable
    [Documentation]    Routes to remote CIDRs in table 200 should be gone.
    ${stdout}=    Command On Cluster    cluster-a    ${IP_CMD} route show table 200
    FOR    ${cidr}    IN
    ...    ${CLUSTER_B_POD_CIDR}
    ...    ${CLUSTER_B_SVC_CIDR}
    ...    ${CLUSTER_C_POD_CIDR}
    ...    ${CLUSTER_C_SVC_CIDR}
        Should Not Contain    ${stdout}    ${cidr}
    END

No IP Rules For Table 200 After Disable
    [Documentation]    IP rules directing to table 200 should be gone.
    ${stdout}=    Command On Cluster    cluster-a    ${IP_CMD} rule show
    FOR    ${cidr}    IN
    ...    ${CLUSTER_B_POD_CIDR}
    ...    ${CLUSTER_B_SVC_CIDR}
    ...    ${CLUSTER_C_POD_CIDR}
    ...    ${CLUSTER_C_SVC_CIDR}
        Should Not Contain    ${stdout}    to ${cidr} lookup 200
    END

No Service Routes In Table 201 After Disable
    [Documentation]    Service routes in table 201 should be gone.
    ${stdout}=    Command On Cluster    cluster-a    ${IP_CMD} route show table 201
    Should Not Contain    ${stdout}    ${CLUSTER_A_SVC_CIDR}

No Service IP Rules After Disable
    [Documentation]    Service IP rules for table 201 should be gone.
    ${stdout}=    Command On Cluster    cluster-a    ${IP_CMD} rule show
    FOR    ${cidr}    IN
    ...    ${CLUSTER_B_POD_CIDR}
    ...    ${CLUSTER_B_SVC_CIDR}
    ...    ${CLUSTER_C_POD_CIDR}
    ...    ${CLUSTER_C_SVC_CIDR}
        Should Not Contain    ${stdout}    from ${cidr} to ${CLUSTER_A_SVC_CIDR} lookup 201
    END

No NFTables Bypass Rules After Disable
    [Documentation]    C2CC nftables masquerade bypass rules should be gone.
    ${stdout}=    Command On Cluster    cluster-a
    ...    nft list chain inet ovn-kubernetes ovn-kube-pod-subnet-masq
    Should Not Contain    ${stdout}    c2cc-no-masq:

No OVN Static Routes After Disable
    [Documentation]    OVN static routes with microshift-c2cc tag should be gone.
    ${pod}=    Oc On Cluster    cluster-a
    ...    oc get pod -n openshift-ovn-kubernetes -l app=ovnkube-master -o jsonpath='{.items[0].metadata.name}'
    ${stdout}=    Oc On Cluster
    ...    cluster-a
    ...    oc exec -n openshift-ovn-kubernetes ${pod} -- ovn-nbctl find Logical_Router_Static_Route external_ids:k8s.ovn.org/owner-controller=microshift-c2cc
    Should Not Contain    ${stdout}    microshift-c2cc

No Node SNAT Annotation After Disable
    [Documentation]    The SNAT-exclude annotation should be absent.
    ${stdout}=    Oc On Cluster    cluster-a
    ...    oc get node -o jsonpath='{.items[0].metadata.annotations.k8s\\.ovn\\.org/node-ingress-snat-exclude-subnets}'
    Should Be Empty    ${stdout}

No C2CC Tracking Annotation After Disable
    [Documentation]    The C2CC tracking annotation should be absent.
    ${stdout}=    Oc On Cluster    cluster-a
    ...    oc get node -o jsonpath='{.items[0].metadata.annotations.microshift\\.io/c2cc-snat-subnets}'
    Should Be Empty    ${stdout}

No Dual Stack Linux Routes In Table 200 After Disable
    [Documentation]    Dual-stack routes to remote CIDRs in table 200 should be gone.
    Skip If    '${CLUSTER_A_POD_CIDR_DUAL}' == ''    Dual-stack CIDRs not configured
    ${ip_cmd}=    IP Command For CIDR    ${CLUSTER_B_POD_CIDR_DUAL}
    ${stdout}=    Command On Cluster    cluster-a    ${ip_cmd} route show table 200
    FOR    ${cidr}    IN
    ...    ${CLUSTER_B_POD_CIDR_DUAL}
    ...    ${CLUSTER_B_SVC_CIDR_DUAL}
    ...    ${CLUSTER_C_POD_CIDR_DUAL}
    ...    ${CLUSTER_C_SVC_CIDR_DUAL}
        Should Not Contain    ${stdout}    ${cidr}
    END

No Dual Stack IP Rules For Table 200 After Disable
    [Documentation]    Dual-stack IP rules directing to table 200 should be gone.
    Skip If    '${CLUSTER_A_POD_CIDR_DUAL}' == ''    Dual-stack CIDRs not configured
    ${ip_cmd}=    IP Command For CIDR    ${CLUSTER_B_POD_CIDR_DUAL}
    ${stdout}=    Command On Cluster    cluster-a    ${ip_cmd} rule show
    FOR    ${cidr}    IN
    ...    ${CLUSTER_B_POD_CIDR_DUAL}
    ...    ${CLUSTER_B_SVC_CIDR_DUAL}
    ...    ${CLUSTER_C_POD_CIDR_DUAL}
    ...    ${CLUSTER_C_SVC_CIDR_DUAL}
        Should Not Contain    ${stdout}    to ${cidr} lookup 200
    END

No Dual Stack Service Routes In Table 201 After Disable
    [Documentation]    Dual-stack service routes in table 201 should be gone.
    Skip If    '${CLUSTER_A_POD_CIDR_DUAL}' == ''    Dual-stack CIDRs not configured
    ${ip_cmd}=    IP Command For CIDR    ${CLUSTER_A_SVC_CIDR_DUAL}
    ${stdout}=    Command On Cluster    cluster-a    ${ip_cmd} route show table 201
    Should Not Contain    ${stdout}    ${CLUSTER_A_SVC_CIDR_DUAL}

No Dual Stack Service IP Rules After Disable
    [Documentation]    Dual-stack service IP rules for table 201 should be gone.
    Skip If    '${CLUSTER_A_POD_CIDR_DUAL}' == ''    Dual-stack CIDRs not configured
    ${ip_cmd}=    IP Command For CIDR    ${CLUSTER_B_POD_CIDR_DUAL}
    ${stdout}=    Command On Cluster    cluster-a    ${ip_cmd} rule show
    FOR    ${cidr}    IN
    ...    ${CLUSTER_B_POD_CIDR_DUAL}
    ...    ${CLUSTER_B_SVC_CIDR_DUAL}
    ...    ${CLUSTER_C_POD_CIDR_DUAL}
    ...    ${CLUSTER_C_SVC_CIDR_DUAL}
        Should Not Contain    ${stdout}    from ${cidr} to ${CLUSTER_A_SVC_CIDR_DUAL} lookup 201
    END

C2CC Controller Logged Cleanup
    [Documentation]    The controller should have logged that it is disabled and cleaning up.
    ${stdout}=    Command On Cluster    cluster-a
    ...    journalctl -u microshift --grep "C2CC is disabled" --no-pager -q
    Should Contain    ${stdout}    C2CC is disabled


*** Keywords ***
Setup
    [Documentation]    Register clusters, then disable C2CC on Cluster A and wait for restart.
    C2CC Suite Setup
    Disable C2CC On Cluster    cluster-a

Teardown
    [Documentation]    Re-enable C2CC on Cluster A, then close connections.
    Enable C2CC On Cluster    cluster-a
    C2CC Suite Teardown

Disable C2CC On Cluster
    [Documentation]    Move the C2CC config drop-in aside and restart MicroShift.
    [Arguments]    ${alias}
    Command On Cluster    ${alias}    mv ${C2CC_CONFIG_PATH} ${C2CC_CONFIG_PATH}.bak
    Command On Cluster    ${alias}    systemctl restart microshift
    Wait Until Keyword Succeeds    ${CLEANUP_TIMEOUT}    ${CLEANUP_RETRY}
    ...    Verify Cluster Is Healthy    ${alias}

Enable C2CC On Cluster
    [Documentation]    Restore the C2CC config drop-in and restart MicroShift.
    [Arguments]    ${alias}
    Command On Cluster    ${alias}    mv ${C2CC_CONFIG_PATH}.bak ${C2CC_CONFIG_PATH}
    Command On Cluster    ${alias}    systemctl restart microshift
    Wait Until Keyword Succeeds    ${CLEANUP_TIMEOUT}    ${CLEANUP_RETRY}
    ...    Verify Cluster Is Healthy    ${alias}
