*** Settings ***
Documentation       Disruptive/resilience tests for C2CC.
...                 Each test injects a disruption (service restart, pod deletion,
...                 NIC outage), waits for recovery, and verifies that C2CC
...                 infrastructure and cross-cluster connectivity are restored.

Resource            ../../../resources/microshift-process.resource
Resource            ../../../resources/kubeconfig.resource
Resource            ../../../resources/oc.resource
Resource            ../../../resources/c2cc.resource

Suite Setup         C2CC Suite Setup    deploy_workloads=${TRUE}
Suite Teardown      C2CC Suite Teardown    cleanup_workloads=${TRUE}

Test Tags           disruptive


*** Variables ***
${RECOVERY_TIMEOUT}     5m
${RECOVERY_RETRY}       15s
${HOST2_VM_NAME}        ${EMPTY}
${HOST3_VM_NAME}        ${EMPTY}
${DISABLED_VM}          ${EMPTY}
@{DISABLED_IFACES}      @{EMPTY}


*** Test Cases ***
Recovery After MicroShift Restart On One Cluster
    [Documentation]    Restart microshift.service on cluster-b.
    ...    Verify infrastructure and connectivity recover.
    [Setup]    Verify All RemoteClusters Healthy
    Command On Cluster    cluster-b    systemctl restart microshift
    Verify All Clusters Are Healthy

Recovery After OVN-K Pod Restart On One Cluster
    [Documentation]    Force-delete all OVN-K pods on cluster-a.
    ...    Verify OVN-K pods recover and C2CC state is restored.
    [Setup]    Verify All RemoteClusters Healthy
    Oc On Cluster    cluster-a
    ...    oc delete pods -n openshift-ovn-kubernetes --all --force --grace-period=0
    ...    allow_fail=${TRUE}
    Wait For OVN-K Pods Ready On Cluster    cluster-a
    Verify All Clusters Are Healthy

Recovery After NetworkManager Restart On One Cluster
    [Documentation]    Restart NetworkManager on cluster-c.
    ...    Verify kernel routes/rules are restored and connectivity recovers.
    [Setup]    Verify All RemoteClusters Healthy
    Command On Cluster    cluster-c    systemctl restart NetworkManager
    Verify All Clusters Are Healthy

Recovery After NIC Outage On One Cluster
    [Documentation]    Disable then re-enable NICs on cluster-b via virsh.
    ...    Verify SSH reconnection, infrastructure, and connectivity.
    [Setup]    Verify All RemoteClusters Healthy
    ${vnet_ifaces}    Disable All NICs For VM    ${HOST2_VM_NAME}
    VAR    ${DISABLED_VM}    ${HOST2_VM_NAME}    scope=TEST
    VAR    @{DISABLED_IFACES}    @{vnet_ifaces}    scope=TEST
    Verify RemoteCluster Unhealthy On Observers    ${HOST2_IP}    cluster-a    cluster-c
    ...    disrupted_ipv6=${HOST2_IPV6}
    Enable All NICs For VM    ${HOST2_VM_NAME}    ${vnet_ifaces}
    Reconnect To Cluster    cluster-b    ${HOST2_IP}    ${HOST2_SSH_PORT}    ${KUBECONFIG_B}
    ...    timeout=${RECOVERY_TIMEOUT}
    VAR    ${DISABLED_VM}    ${EMPTY}    scope=TEST
    Verify All Clusters Are Healthy
    [Teardown]    Restore NICs And Reconnect
    ...    ${HOST2_VM_NAME}    cluster-b    ${HOST2_IP}    ${HOST2_SSH_PORT}    ${KUBECONFIG_B}

Recovery After MicroShift Restart On Two Clusters
    [Documentation]    Restart microshift.service on cluster-a and cluster-c simultaneously.
    ...    Verify both clusters recover independently.
    [Setup]    Verify All RemoteClusters Healthy
    Disruptive Command On Cluster    cluster-a    nohup systemctl restart microshift &>/dev/null &
    Disruptive Command On Cluster    cluster-c    nohup systemctl restart microshift &>/dev/null &
    Sleep    10s
    Verify All Clusters Are Healthy

Recovery After MicroShift Restart On All Clusters
    [Documentation]    Restart microshift.service on all three clusters simultaneously.
    ...    Verify all clusters recover independently.
    [Setup]    Verify All RemoteClusters Healthy
    Disruptive Command On Cluster    cluster-a    nohup systemctl restart microshift &>/dev/null &
    Disruptive Command On Cluster    cluster-b    nohup systemctl restart microshift &>/dev/null &
    Disruptive Command On Cluster    cluster-c    nohup systemctl restart microshift &>/dev/null &
    Sleep    10s
    Verify All Clusters Are Healthy

Recovery After OVN-K Restart On One Cluster And NIC Outage On Another Cluster
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
    ...    disrupted_ipv6=${HOST3_IPV6}
    Enable All NICs For VM    ${HOST3_VM_NAME}    ${vnet_ifaces}
    Wait For OVN-K Pods Ready On Cluster    cluster-b
    Reconnect To Cluster    cluster-c    ${HOST3_IP}    ${HOST3_SSH_PORT}    ${KUBECONFIG_C}
    ...    timeout=${RECOVERY_TIMEOUT}
    VAR    ${DISABLED_VM}    ${EMPTY}    scope=TEST
    Verify All Clusters Are Healthy
    [Teardown]    Restore NICs And Reconnect
    ...    ${HOST3_VM_NAME}    cluster-c    ${HOST3_IP}    ${HOST3_SSH_PORT}    ${KUBECONFIG_C}

Recovery After NM Restart On One Cluster And MicroShift Restart On Another Cluster
    [Documentation]    Restart NetworkManager on cluster-a and microshift on cluster-b.
    ...    Verify both clusters recover from different service disruptions.
    [Setup]    Verify All RemoteClusters Healthy
    Disruptive Command On Cluster    cluster-a    nohup systemctl restart NetworkManager &>/dev/null &
    Disruptive Command On Cluster    cluster-b    nohup systemctl restart microshift &>/dev/null &
    Sleep    10s
    Verify All Clusters Are Healthy


*** Keywords ***
Verify All Clusters Are Healthy
    [Documentation]    Wait for each cluster's healthcheck, then verify full C2CC recovery.
    Verify All Clusters Healthy    timeout=${RECOVERY_TIMEOUT}    retry=${RECOVERY_RETRY}
    Verify Full C2CC Stack

Restore NICs And Reconnect
    [Documentation]    Re-enable NICs if they were left disabled by a failed NIC-outage test.
    ...    Safe to call even when NICs are already up (Enable is idempotent via virsh).
    [Arguments]    ${vm_name}    ${alias}    ${host_ip}    ${ssh_port}    ${kubeconfig}
    IF    '${DISABLED_VM}' != ''
        Enable All NICs For VM    ${vm_name}    ${DISABLED_IFACES}
        Reconnect To Cluster    ${alias}    ${host_ip}    ${ssh_port}    ${kubeconfig}
    END
