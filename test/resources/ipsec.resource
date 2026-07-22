*** Settings ***
Documentation       Keywords for IPsec (Libreswan) operations in C2CC test scenarios.
...                 Provides tunnel verification, XFRM counter checks, ESP packet capture,
...                 SA management, and nftables IPsec enforcement helpers.

Library             String
Resource            c2cc.resource


*** Variables ***
${IPSEC_SA_TIMEOUT}     60s
${IPSEC_SA_RETRY}       5s


*** Keywords ***
Verify IPsec Tunnel Is Established
    [Documentation]    Verify that at least one IPsec tunnel is established on the given cluster.
    [Arguments]    ${alias}
    ${stdout}=    Command On Cluster    ${alias}    ipsec trafficstatus
    Should Contain    ${stdout}    type=ESP

Verify All IPsec Tunnels On Cluster
    [Documentation]    Verify that the expected number of IPsec tunnels are established.
    [Arguments]    ${alias}    ${expected_count}=2
    ${stdout}=    Command On Cluster    ${alias}    ipsec trafficstatus
    ${count}=    Get Count    ${stdout}    type=ESP
    Should Be True    ${count} >= ${expected_count}
    ...    Expected ${expected_count} IPsec tunnels but found ${count}

Stop IPsec Service On Cluster
    [Documentation]    Stop the ipsec systemd service on the given cluster.
    [Arguments]    ${alias}
    Disruptive Command On Cluster    ${alias}    systemctl stop ipsec

Start IPsec Service On Cluster
    [Documentation]    Start the ipsec systemd service on the given cluster.
    [Arguments]    ${alias}
    Command On Cluster    ${alias}    systemctl start ipsec

Restart IPsec Service On Cluster
    [Documentation]    Restart the ipsec systemd service on the given cluster.
    [Arguments]    ${alias}
    Command On Cluster    ${alias}    systemctl restart ipsec

Wait For IPsec Tunnel Reestablishment
    [Documentation]    Wait for IPsec tunnels to be re-established after disruption.
    [Arguments]    ${alias}    ${expected_count}=2
    Wait Until Keyword Succeeds    ${IPSEC_SA_TIMEOUT}    ${IPSEC_SA_RETRY}
    ...    Verify All IPsec Tunnels On Cluster    ${alias}    ${expected_count}

Get XFRM Byte Counters
    [Documentation]    Return the total XFRM SA byte counter (sum across all ESP SAs).
    [Arguments]    ${alias}
    ${stdout}=    Command On Cluster    ${alias}
    ...    ip -s xfrm state | awk '/bytes/{gsub(/[^0-9]/,"",$1); sum+=$1} END{print sum+0}'
    RETURN    ${stdout}

Verify XFRM Counters Incremented
    [Documentation]    Verify XFRM byte counters have incremented from a baseline value.
    [Arguments]    ${alias}    ${baseline_bytes}
    ${current_bytes}=    Get XFRM Byte Counters    ${alias}
    ${baseline_int}=    Convert To Integer    ${baseline_bytes}
    ${current_int}=    Convert To Integer    ${current_bytes}
    Should Be True    ${current_int} > ${baseline_int}
    ...    XFRM byte counter did not increment: baseline=${baseline_int}, current=${current_int}

Start Tcpdump For ESP On Cluster
    [Documentation]    Start tcpdump in the background capturing ESP packets.
    ...    Returns the pcap file path.
    [Arguments]    ${alias}    ${iface}=enp1s0
    VAR    ${pcap_file}=    /tmp/esp-capture-${alias}.pcap
    Disruptive Command On Cluster    ${alias}    rm -f ${pcap_file}
    ${conn_id}=    Get From Dictionary    ${C2CC_SSH_IDS}    ${alias}
    SSHLibrary.Switch Connection    ${conn_id}
    SSHLibrary.Start Command    tcpdump -U -i ${iface} -w ${pcap_file} esp    sudo=True
    Sleep    2s    reason=Wait for tcpdump to initialize packet capture
    RETURN    ${pcap_file}

Wait For Tcpdump And Verify ESP
    [Documentation]    Stop tcpdump, then verify ESP packets in pcap.
    [Arguments]    ${alias}    ${pcap_file}
    Sleep    3s    reason=Wait for ESP packets to be captured
    Disruptive Command On Cluster    ${alias}    pkill -INT -f "tcpdump.*${pcap_file}"
    ${conn_id}=    Get From Dictionary    ${C2CC_SSH_IDS}    ${alias}
    SSHLibrary.Switch Connection    ${conn_id}
    ${stdout}    ${stderr}=    SSHLibrary.Read Command Output    return_stderr=True
    Log    tcpdump stdout: ${stdout}    level=DEBUG
    Log    tcpdump stderr: ${stderr}    level=DEBUG
    ${stdout}=    Command On Cluster    ${alias}
    ...    tcpdump -r ${pcap_file} 2>/dev/null | head -5
    Should Contain    ${stdout}    ESP
    Disruptive Command On Cluster    ${alias}    rm -f ${pcap_file}

Add NFTables IPsec Enforcement Rules
    [Documentation]    Add nftables rules that drop non-IPsec traffic destined for local pod CIDRs.
    ...    Uses 'meta ipsec missing' match (kernel 5.10+) to detect unencrypted packets.
    ...    In tunnel mode, decrypted packets have pod/service CIDRs as dst — plaintext
    ...    packets arriving without ESP are caught by this rule.
    ...    The chain hooks into forward: inbound cross-cluster packets are not addressed
    ...    to the host but forwarded into the OVN network (local gateway mode), so an
    ...    input-hook chain would never see them.
    [Arguments]    ${alias}    ${local_pod_cidr}
    Command On Cluster    ${alias}    nft add table inet c2cc_ipsec_test
    Command On Cluster    ${alias}
    ...    nft 'add chain inet c2cc_ipsec_test enforce { type filter hook forward priority -150; policy accept; }'
    Command On Cluster    ${alias}
    ...    nft add rule inet c2cc_ipsec_test enforce ip daddr ${local_pod_cidr} meta ipsec missing counter drop

Remove NFTables IPsec Enforcement Rules
    [Documentation]    Remove C2CC IPsec enforcement nftables table and all its rules.
    [Arguments]    ${alias}
    Disruptive Command On Cluster
    ...    ${alias}
    ...    nft delete table inet c2cc_ipsec_test 2>/dev/null || true

Curl Should Fail From Cluster
    [Documentation]    Verify curl from curl-pod times out or fails (no "Hello from" in response).
    [Arguments]    ${alias}    ${ip}    ${port}    ${ns}
    ${stdout}=    Oc On Cluster    ${alias}
    ...    oc exec curl-pod -n ${ns} -- curl -sS --max-time 5 http://${ip}:${port}/cgi-bin/hello 2>&1 || true
    ...    allow_fail=${TRUE}
    Should Not Contain    ${stdout}    Hello from

Curl From Host Should Fail
    [Documentation]    Curl directly from the host (not from a pod) to a remote pod IP.
    ...    Expects failure — host-originated traffic is not matched by tunnel-mode
    ...    IPsec selectors scoped to pod/service CIDRs.
    [Arguments]    ${alias}    ${pod_ip}    ${port}
    ${stdout}=    Disruptive Command On Cluster    ${alias}
    ...    curl -sS --max-time 5 http://${pod_ip}:${port}/cgi-bin/hello 2>&1 || true
    Should Not Contain    ${stdout}    Hello from
