*** Settings ***
Documentation       IPsec E2E tests for C2CC.
...                 Validates that C2CC cross-cluster connectivity works through a
...                 Libreswan tunnel-mode IPsec mesh (subnet selectors, no Geneve).
...
...                 Tests cover ESP encapsulation, connectivity through the tunnel,
...                 source IP preservation, policy enforcement (SA flush/restore),
...                 plaintext rejection, host-to-pod rejection, and MTU validation.

Resource            ../../resources/microshift-process.resource
Resource            ../../resources/kubeconfig.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/c2cc.resource
Resource            ../../resources/ipsec.resource

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           c2cc    ipsec


*** Variables ***
&{NAMESPACES}       cluster-a=${EMPTY}    cluster-b=${EMPTY}    cluster-c=${EMPTY}


*** Test Cases ***
IPsec Tunnels Established On All Clusters
    [Documentation]    Verify all 3 hosts have 8 IPsec tunnel SAs each.
    ...    Each host has 2 remote hosts x 4 subnet pairs (2 local x 2 remote CIDRs).
    Verify All IPsec Tunnels On Cluster    cluster-a    expected_count=8
    Verify All IPsec Tunnels On Cluster    cluster-b    expected_count=8
    Verify All IPsec Tunnels On Cluster    cluster-c    expected_count=8

ESP Encapsulation On Wire
    [Documentation]    Capture packets on the wire and verify ESP encapsulation.
    ...    Records XFRM byte counters before and after traffic, captures ESP packets
    ...    via tcpdump, and verifies counters incremented.
    ${baseline_a}=    Get XFRM Byte Counters    cluster-a
    ${baseline_b}=    Get XFRM Byte Counters    cluster-b

    ${pcap_file}=    Start Tcpdump For ESP On Cluster    cluster-b
    ${ip_dest}=    Get Hello Pod IP    cluster-b
    Curl From Cluster    cluster-a    ${ip_dest}    8080
    Wait For Tcpdump And Verify ESP    cluster-b    ${pcap_file}

    Verify XFRM Counters Incremented    cluster-a    ${baseline_a}
    Verify XFRM Counters Incremented    cluster-b    ${baseline_b}

Cross Cluster Connectivity Through IPsec
    [Documentation]    Verify pods on all clusters can reach pods/services on all
    ...    other clusters through the IPsec tunnel.
    [Template]    Test Connectivity Between Clusters
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
    [Template]    Test Source IP Preserved Between Clusters
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

Near MTU Packet Through IPsec Tunnel
    [Documentation]    Send near-MTU-sized payloads through IPsec tunnel-mode
    ...    encapsulation.    Verifies no MTU blackhole from DF-bit issues.
    ...    ESP overhead ~36-52B total.
    ${ip_dest}=    Get Hello Pod IP    cluster-b
    Send Large Payload And Verify    cluster-a    ${ip_dest}    1300
    Send Large Payload And Verify    cluster-a    ${ip_dest}    1400


*** Keywords ***
Setup
    [Documentation]    Register all clusters, verify IPsec tunnels, deploy test workloads.
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig
    Register Local Cluster    cluster-a
    Register Remote Cluster    cluster-b    ${HOST2_IP}    ${HOST2_SSH_PORT}    ${KUBECONFIG_B}
    Register Remote Cluster    cluster-c    ${HOST3_IP}    ${HOST3_SSH_PORT}    ${KUBECONFIG_C}
    Deploy Test Workloads

Teardown
    [Documentation]    Remove test workloads, ensure IPsec is running, close connections.
    Cleanup Test Workloads
    Ensure IPsec Running On All Clusters
    Teardown All Remote Clusters
    Remove Kubeconfig
    Logout MicroShift Host

Ensure IPsec Running On All Clusters
    [Documentation]    Make sure ipsec service is running on all clusters.
    ...    Tests may have stopped it.
    FOR    ${alias}    IN    cluster-a    cluster-b    cluster-c
        Start IPsec Service On Cluster    ${alias}
    END

Restore IPsec And Verify
    [Documentation]    Restart ipsec and wait for tunnels to come back up.
    [Arguments]    ${alias}
    Restart IPsec Service On Cluster    ${alias}
    Wait For IPsec Tunnel Reestablishment    ${alias}    expected_count=8

Restore IPsec With Enforcement Cleanup
    [Documentation]    Remove enforcement rules and restore IPsec.
    [Arguments]    ${ipsec_alias}    ${enforcement_alias}
    Remove NFTables IPsec Enforcement Rules    ${enforcement_alias}
    Start IPsec Service On Cluster    ${ipsec_alias}
    Wait For IPsec Tunnel Reestablishment    ${ipsec_alias}    expected_count=8

Get Hello Pod IP
    [Documentation]    Get the pod IP of hello-microshift on the given cluster.
    [Arguments]    ${alias}
    ${ip}=    Oc On Cluster    ${alias}
    ...    oc get pod hello-microshift -n ${NAMESPACES}[${alias}] -o jsonpath='{.status.podIP}'
    RETURN    ${ip}

Get Hello Service IP
    [Documentation]    Get the ClusterIP of the hello-microshift service on the given cluster.
    [Arguments]    ${alias}
    ${ip}=    Oc On Cluster    ${alias}
    ...    oc get svc hello-microshift -n ${NAMESPACES}[${alias}] -o jsonpath='{.spec.clusterIP}'
    RETURN    ${ip}

Get Curl Pod IP
    [Documentation]    Get the pod IP of curl-pod on the given cluster.
    [Arguments]    ${alias}
    ${ip}=    Oc On Cluster    ${alias}
    ...    oc get pod curl-pod -n ${NAMESPACES}[${alias}] -o jsonpath='{.status.podIP}'
    RETURN    ${ip}

Curl From Cluster
    [Documentation]    Exec curl from curl-pod on the given cluster.
    [Arguments]    ${alias}    ${ip}    ${port}
    ${stdout}=    Oc On Cluster    ${alias}
    ...    oc exec curl-pod -n ${NAMESPACES}[${alias}] -- curl -sS --max-time 10 http://${ip}:${port}/cgi-bin/hello
    RETURN    ${stdout}

Test Connectivity Between Clusters
    [Documentation]    Verify pod on source can reach endpoint on destination.
    [Arguments]    ${source}    ${destination}    ${endpoint_type}
    IF    '${endpoint_type}' == 'pod'
        ${ip_dest}=    Get Hello Pod IP    ${destination}
    ELSE IF    '${endpoint_type}' == 'service'
        ${ip_dest}=    Get Hello Service IP    ${destination}
    ELSE
        Fail    Invalid endpoint_type: ${endpoint_type}. Must be 'pod' or 'service'.
    END
    ${stdout}=    Curl From Cluster    ${source}    ${ip_dest}    8080
    Should Contain    ${stdout}    Hello from

Test Source IP Preserved Between Clusters
    [Documentation]    Verify source pod IP is preserved through IPsec tunnel (no SNAT).
    [Arguments]    ${source}    ${destination}    ${endpoint_type}
    ${curl_pod_ip}=    Get Curl Pod IP    ${source}
    IF    '${endpoint_type}' == 'pod'
        ${ip_dest}=    Get Hello Pod IP    ${destination}
    ELSE IF    '${endpoint_type}' == 'service'
        ${ip_dest}=    Get Hello Service IP    ${destination}
    ELSE
        Fail    Invalid endpoint_type: ${endpoint_type}. Must be 'pod' or 'service'.
    END
    ${stdout}=    Curl From Cluster    ${source}    ${ip_dest}    8080
    Should Contain    ${stdout}    source: ${curl_pod_ip}

Send Large Payload And Verify
    [Documentation]    Send a large payload via curl POST and verify it succeeds.
    [Arguments]    ${alias}    ${ip}    ${size}
    ${stdout}=    Oc On Cluster
    ...    ${alias}
    ...    oc exec curl-pod -n ${NAMESPACES}[${alias}] -- sh -c 'dd if=/dev/zero bs=${size} count=1 2>/dev/null | curl -sS --max-time 15 --data-binary @- http://${ip}:8080/cgi-bin/hello'
    Should Contain    ${stdout}    Hello from
