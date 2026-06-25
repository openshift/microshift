*** Settings ***
Documentation       Chaos/resilience tests for C2CC.
...                 Each test injects a disruption (service restart, pod deletion,
...                 NIC outage), waits for recovery, and verifies that C2CC
...                 infrastructure and cross-cluster connectivity are restored.

Resource            ../../resources/microshift-process.resource
Resource            ../../resources/kubeconfig.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/c2cc.resource

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           chaos


*** Variables ***
${RECOVERY_TIMEOUT}         5m
${RECOVERY_RETRY}           15s
${INFRA_VERIFY_TIMEOUT}     3m
${INFRA_VERIFY_RETRY}       10s
${HOST2_VM_NAME}            ${EMPTY}
${HOST3_VM_NAME}            ${EMPTY}
${DISABLED_VM}              ${EMPTY}
@{DISABLED_IFACES}          @{EMPTY}


*** Test Cases ***
Recovery After MicroShift Restart On One Cluster
    [Documentation]    Restart microshift.service on cluster-b.
    ...    Verify infrastructure and connectivity recover.
    [Setup]    Verify All RemoteClusters Healthy
    Command On Cluster    cluster-b    systemctl restart microshift
    Verify Clusters Are Healthy    cluster-b

Recovery After OVN-K Pod Restart On Cluster A
    [Documentation]    Force-delete all OVN-K pods on cluster-a.
    ...    Verify OVN-K pods recover and C2CC state is restored.
    [Setup]    Verify All RemoteClusters Healthy
    Oc On Cluster    cluster-a
    ...    oc delete pods -n openshift-ovn-kubernetes --all --force --grace-period=0
    ...    allow_fail=${TRUE}
    Wait For OVN-K Pods Ready On Cluster    cluster-a
    Verify Clusters Are Healthy    cluster-a

Recovery After NetworkManager Restart On Cluster C
    [Documentation]    Restart NetworkManager on cluster-c.
    ...    Verify kernel routes/rules are restored and connectivity recovers.
    [Setup]    Verify All RemoteClusters Healthy
    Command On Cluster    cluster-c    systemctl restart NetworkManager
    Verify Clusters Are Healthy    cluster-c

Recovery After NIC Outage On Cluster B
    [Documentation]    Disable then re-enable NICs on cluster-b via virsh.
    ...    Verify SSH reconnection, infrastructure, and connectivity.
    [Setup]    Verify All RemoteClusters Healthy
    ${vnet_ifaces}    Disable All NICs For VM    ${HOST2_VM_NAME}
    VAR    ${DISABLED_VM}    ${HOST2_VM_NAME}    scope=TEST
    VAR    @{DISABLED_IFACES}    @{vnet_ifaces}    scope=TEST
    Verify RemoteCluster Unhealthy On Observers    ${HOST2_IP}    cluster-a    cluster-c
    Enable All NICs For VM    ${HOST2_VM_NAME}    ${vnet_ifaces}
    Reconnect To Cluster    cluster-b    ${HOST2_IP}    ${HOST2_SSH_PORT}    ${KUBECONFIG_B}
    ...    timeout=${RECOVERY_TIMEOUT}
    VAR    ${DISABLED_VM}    ${EMPTY}    scope=TEST
    Verify Clusters Are Healthy    cluster-b
    [Teardown]    Restore NICs And Reconnect
    ...    ${HOST2_VM_NAME}    cluster-b    ${HOST2_IP}    ${HOST2_SSH_PORT}    ${KUBECONFIG_B}

Recovery After MicroShift Restart On Clusters A And C
    [Documentation]    Restart microshift.service on cluster-a and cluster-c simultaneously.
    ...    Verify both clusters recover independently.
    [Setup]    Verify All RemoteClusters Healthy
    Disruptive Command On Cluster    cluster-a    nohup systemctl restart microshift &>/dev/null &
    Disruptive Command On Cluster    cluster-c    nohup systemctl restart microshift &>/dev/null &
    Sleep    10s
    Verify Clusters Are Healthy    cluster-a    cluster-c

Recovery After MicroShift Restart On All Clusters
    [Documentation]    Restart microshift.service on all three clusters simultaneously.
    ...    Verify all clusters recover independently.
    [Setup]    Verify All RemoteClusters Healthy
    Disruptive Command On Cluster    cluster-a    nohup systemctl restart microshift &>/dev/null &
    Disruptive Command On Cluster    cluster-b    nohup systemctl restart microshift &>/dev/null &
    Disruptive Command On Cluster    cluster-c    nohup systemctl restart microshift &>/dev/null &
    Sleep    10s
    Verify Clusters Are Healthy    cluster-a    cluster-b    cluster-c

Recovery After OVN-K Restart On B And NIC Outage On C
    [Documentation]    Delete OVN-K pods on cluster-b and disable NICs on cluster-c.
    ...    Verify both clusters recover from different failure modes.
    [Setup]    Verify All RemoteClusters Healthy
    Oc On Cluster    cluster-b
    ...    oc delete pods -n openshift-ovn-kubernetes --all --force --grace-period=0
    ...    allow_fail=${TRUE}
    ${vnet_ifaces}    Disable All NICs For VM    ${HOST3_VM_NAME}
    VAR    ${DISABLED_VM}    ${HOST3_VM_NAME}    scope=TEST
    VAR    @{DISABLED_IFACES}    @{vnet_ifaces}    scope=TEST
    Verify RemoteCluster Unhealthy On Observers    ${HOST3_IP}    cluster-a
    Enable All NICs For VM    ${HOST3_VM_NAME}    ${vnet_ifaces}
    Wait For OVN-K Pods Ready On Cluster    cluster-b
    Reconnect To Cluster    cluster-c    ${HOST3_IP}    ${HOST3_SSH_PORT}    ${KUBECONFIG_C}
    ...    timeout=${RECOVERY_TIMEOUT}
    VAR    ${DISABLED_VM}    ${EMPTY}    scope=TEST
    Verify Clusters Are Healthy    cluster-b    cluster-c
    [Teardown]    Restore NICs And Reconnect
    ...    ${HOST3_VM_NAME}    cluster-c    ${HOST3_IP}    ${HOST3_SSH_PORT}    ${KUBECONFIG_C}

Recovery After NM Restart On A And MicroShift Restart On B
    [Documentation]    Restart NetworkManager on cluster-a and microshift on cluster-b.
    ...    Verify both clusters recover from different service disruptions.
    [Setup]    Verify All RemoteClusters Healthy
    Disruptive Command On Cluster    cluster-a    nohup systemctl restart NetworkManager &>/dev/null &
    Disruptive Command On Cluster    cluster-b    nohup systemctl restart microshift &>/dev/null &
    Sleep    10s
    Verify Clusters Are Healthy    cluster-a    cluster-b


*** Keywords ***
Verify Clusters Are Healthy
    [Documentation]    Wait for each cluster's healthcheck, then verify full C2CC recovery.
    [Arguments]    @{clusters}
    FOR    ${cluster}    IN    @{clusters}
        Wait Until Keyword Succeeds    ${RECOVERY_TIMEOUT}    ${RECOVERY_RETRY}
        ...    Verify Cluster Is Healthy    ${cluster}
    END
    Verify Full Recovery On Clusters    @{clusters}

Verify Full Recovery On Clusters
    [Documentation]    Wait for RemoteCluster CRs to converge, verify C2CC infrastructure
    ...    on each specified cluster, then verify cross-cluster connectivity and DNS.
    [Arguments]    @{clusters}
    Verify All RemoteClusters Healthy
    FOR    ${cluster}    IN    @{clusters}
        Wait Until Keyword Succeeds    ${INFRA_VERIFY_TIMEOUT}    ${INFRA_VERIFY_RETRY}
        ...    Verify C2CC Infrastructure On Cluster    ${cluster}
    END
    Verify Cross Cluster Connectivity
    Verify Cross Cluster DNS

Restore NICs And Reconnect
    [Documentation]    Re-enable NICs if they were left disabled by a failed NIC-outage test.
    ...    Safe to call even when NICs are already up (Enable is idempotent via virsh).
    [Arguments]    ${vm_name}    ${alias}    ${host_ip}    ${ssh_port}    ${kubeconfig}
    IF    '${DISABLED_VM}' != ''
        Enable All NICs For VM    ${vm_name}    ${DISABLED_IFACES}
        Reconnect To Cluster    ${alias}    ${host_ip}    ${ssh_port}    ${kubeconfig}
    END

Setup
    [Documentation]    Set up clusters and deploy test workloads.
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
