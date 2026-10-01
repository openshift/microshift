*** Settings ***
Documentation       Verify C2CC survives VM reboots in escalating scenarios.
...                 Tests single cluster, two clusters simultaneously, and all three
...                 clusters simultaneously. After each reboot cycle, full-stack
...                 verification confirms connectivity, infrastructure, health probes,
...                 and DNS all recover.

Resource            ../../../resources/c2cc.resource

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           disruptive


*** Test Cases ***
Reboot Single Cluster
    [Documentation]    Reboot cluster-a while cluster-b and cluster-c remain up.
    ...    Verifies that the rebooted cluster re-establishes C2CC connectivity
    ...    with both peers.
    [Setup]    Verify All RemoteClusters Healthy
    Reboot Clusters Simultaneously    cluster-a
    Wait For Clusters Ready
    Verify Full C2CC Stack

Reboot Two Clusters Simultaneously
    [Documentation]    Reboot cluster-b and cluster-c at the same time.
    ...    The surviving cluster-a must wait for both peers to recover.
    ...    The two rebooted clusters must also reconnect with each other.
    [Setup]    Verify All RemoteClusters Healthy
    Reboot Clusters Simultaneously    cluster-b    cluster-c
    Wait For Clusters Ready
    Verify Full C2CC Stack

Reboot All Three Clusters Simultaneously
    [Documentation]    Reboot all three clusters at once.
    ...    Every cluster starts from scratch simultaneously - no running peer
    ...    to reference. All must independently reconstruct C2CC state.
    [Setup]    Verify All RemoteClusters Healthy
    Reboot Clusters Simultaneously    cluster-a    cluster-b    cluster-c
    Wait For Clusters Ready
    Verify Full C2CC Stack


*** Keywords ***
Setup
    [Documentation]    Set up clusters, deploy test workloads, and verify full stack.
    C2CC Suite Setup    deploy_workloads=${TRUE}
    Verify Full C2CC Stack

Teardown
    [Documentation]    Remove test workloads and close connections.
    C2CC Suite Teardown    cleanup_workloads=${TRUE}

Wait For Clusters Ready
    [Documentation]    Wait for test pods and service endpoints to become ready
    ...    after a reboot cycle.
    Wait For Test Pods
    Wait For Service Endpoints
