*** Settings ***
Documentation       IPsec E2E tests for C2CC.
...                 Validates that C2CC cross-cluster connectivity works through a
...                 Libreswan tunnel-mode IPsec mesh (subnet selectors, no Geneve).
...
...                 Tests cover ESP encapsulation, connectivity through the tunnel,
...                 source IP preservation, policy enforcement (SA flush/restore),
...                 plaintext rejection, host-to-pod rejection, MTU boundary (1500),
...                 and tunnel recovery. Jumbo MTU tests are in the c2cc-ipsec-mtu scenario.

Resource            ../../../resources/microshift-process.resource
Resource            ../../../resources/kubeconfig.resource
Resource            ../../../resources/oc.resource
Resource            ../../../resources/c2cc.resource
Resource            ../../../resources/ipsec.resource

Suite Setup         C2CC Suite Setup    deploy_workloads=${TRUE}
Suite Teardown      Teardown

Test Tags           c2cc    ipsec


*** Test Cases ***
IPsec Tunnels Established On All Clusters
    [Documentation]    Verify all 3 hosts have 8 IPsec tunnel SAs each.
    ...    Each host has 2 remote hosts x 4 subnet pairs (2 local x 2 remote CIDRs).
    Verify All IPsec Tunnels On Cluster    cluster-a    expected_count=8
    Verify All IPsec Tunnels On Cluster    cluster-b    expected_count=8
    Verify All IPsec Tunnels On Cluster    cluster-c    expected_count=8

ESP Encapsulation On Wire
    [Documentation]    Verify all cross-cluster traffic is ESP-encapsulated by
    ...    capturing packets via tcpdump and checking XFRM counters.
    ${pcap_file}=    Start Tcpdump For ESP On Cluster    cluster-b
    Verify ESP Encapsulation On All Clusters
    Wait For Tcpdump And Verify ESP    cluster-b    ${pcap_file}

Cross Cluster Connectivity Through IPsec
    [Documentation]    Verify pods on all clusters can reach pods/services on all
    ...    other clusters through the IPsec tunnel.
    [Template]    Verify Connectivity Between Clusters
    cluster-a    cluster-b    pod
    cluster-a    cluster-b    service
    cluster-a    cluster-c    pod
    cluster-a    cluster-c    service
    cluster-b    cluster-a    pod
    cluster-b    cluster-a    service
    cluster-b    cluster-c    pod
    cluster-b    cluster-c    service
    cluster-c    cluster-a    pod
    cluster-c    cluster-a    service
    cluster-c    cluster-b    pod
    cluster-c    cluster-b    service

Source IP Preserved Through IPsec
    [Documentation]    Verify cross-cluster traffic through IPsec preserves the
    ...    source pod IP (no SNAT).
    [Template]    Verify Source IP Preserved Between Clusters
    cluster-a    cluster-b    pod
    cluster-a    cluster-b    service
    cluster-a    cluster-c    pod
    cluster-a    cluster-c    service
    cluster-b    cluster-a    pod
    cluster-b    cluster-a    service
    cluster-b    cluster-c    pod
    cluster-b    cluster-c    service
    cluster-c    cluster-a    pod
    cluster-c    cluster-a    service
    cluster-c    cluster-b    pod
    cluster-c    cluster-b    service

Plaintext Rejection When IPsec Stopped
    [Documentation]    Stop IPsec on cluster-a. With nftables enforcement rules on
    ...    cluster-b, verify traffic is dropped rather than sent in plaintext.
    [Setup]    Add NFTables IPsec Enforcement Rules    cluster-b    ${CLUSTER_B_POD_CIDR}

    Stop IPsec Service On Cluster    cluster-a
    Sleep    3s    reason=Wait for SAs to expire
    ${ip_dest}=    Get Hello Pod IP    cluster-b
    Curl Should Fail From Cluster    cluster-a    ${ip_dest}    8080    ${NAMESPACES}[cluster-a]

    [Teardown]    Restore IPsec With Enforcement Cleanup    cluster-a    cluster-b

Host To Remote Pod Rejected Without IPsec
    [Documentation]    Curl directly from cluster-a's host to a pod on cluster-b.
    ...    Host-originated traffic is not matched by tunnel-mode IPsec selectors
    ...    scoped to pod/service CIDRs, so it cannot reach the remote pod.
    ${pod_ip}=    Get Hello Pod IP    cluster-b
    Curl From Host Should Fail    cluster-a    ${pod_ip}    8080

MTU Boundary And TCP Transfer Through IPsec
    [Documentation]    Validate DF-bit boundaries at the ESP-adjusted PMTU and
    ...    large TCP transfers through IPsec on all cluster pairs.
    Ping DF Bit Between All Clusters    64
    Ping DF Bit Between All Clusters    1450
    Ping DF Bit Should Fail Between All Clusters    1472
    Large Payload Between All Clusters    1400
    Large Payload Between All Clusters    65536

Cross Cluster DNS Through IPsec
    [Documentation]    Verify pods can resolve and reach services on remote
    ...    clusters via DNS name through the IPsec tunnel.
    [Template]    Verify Remote Service Via DNS
    cluster-a    cluster-b
    cluster-a    cluster-c
    cluster-b    cluster-a
    cluster-b    cluster-c
    cluster-c    cluster-a
    cluster-c    cluster-b

IPsec Tunnel Recovers After Local Restart
    [Documentation]    Restart IPsec on cluster-a and verify tunnels recover
    ...    and cross-cluster connectivity is restored.
    Restart IPsec Service On Cluster    cluster-a
    Wait For IPsec Tunnel Reestablishment    cluster-a    expected_count=8
    Wait For IPsec Tunnel Reestablishment    cluster-b    expected_count=8
    Wait For IPsec Tunnel Reestablishment    cluster-c    expected_count=8
    ${ip_dest}=    Get Hello Pod IP    cluster-b
    Curl From Cluster    cluster-a    ${ip_dest}    8080

IPsec Tunnel Recovers After Remote Stop Start
    [Documentation]    Stop IPsec on cluster-b, start it back, and verify
    ...    tunnels recover and cross-cluster connectivity is restored.
    Stop IPsec Service On Cluster    cluster-b
    Sleep    5s    reason=Wait for tunnel to go down
    Start IPsec Service On Cluster    cluster-b
    Wait For IPsec Tunnel Reestablishment    cluster-a    expected_count=8
    Wait For IPsec Tunnel Reestablishment    cluster-b    expected_count=8
    Wait For IPsec Tunnel Reestablishment    cluster-c    expected_count=8
    ${ip_dest}=    Get Hello Pod IP    cluster-b
    Curl From Cluster    cluster-a    ${ip_dest}    8080


*** Keywords ***
Teardown
    [Documentation]    Ensure IPsec is running, then remove test workloads and close connections.
    Ensure IPsec Running On All Clusters
    C2CC Suite Teardown    cleanup_workloads=${TRUE}

Ensure IPsec Running On All Clusters
    [Documentation]    Make sure ipsec service is running on all clusters.
    ...    Tests may have stopped it.
    FOR    ${alias}    IN    cluster-a    cluster-b    cluster-c
        Start IPsec Service On Cluster    ${alias}
    END

Restore IPsec With Enforcement Cleanup
    [Documentation]    Remove enforcement rules and restore IPsec.
    [Arguments]    ${ipsec_alias}    ${enforcement_alias}
    Remove NFTables IPsec Enforcement Rules    ${enforcement_alias}
    Start IPsec Service On Cluster    ${ipsec_alias}
    Wait For IPsec Tunnel Reestablishment    ${ipsec_alias}    expected_count=8
