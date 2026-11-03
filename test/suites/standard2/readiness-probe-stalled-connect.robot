*** Settings ***
Documentation       MicroShift must become ready when a readiness probe's first
...                 connect is refused and the fallback connect is never answered
...                 (for example ::1 on a host without IPv6 loopback). Before the
...                 probe bounded its connects, one such attempt blocked for the
...                 kernel's SYN retry schedule, longer than the probe's budget,
...                 and MicroShift stopped itself. The host needs an IPv6 output
...                 route to ::1 (loopback or not) so the black-holed SYN is sent.

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-host.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/systemd.resource
Library             ../../resources/journalctl.py

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           restart    slow


*** Variables ***
${KUBELET_HEALTHZ_PORT}     10248
${NFT_TABLE_V6}             ushift_probe_test_v6
${NFT_TABLE_V4}             ushift_probe_test_v4


*** Test Cases ***
MicroShift Becomes Ready When A Probe Connect Is Never Answered
    [Documentation]    Black-hole IPv6 connects to the kubelet healthz port and
    ...    refuse IPv4 ones, start MicroShift, wait until the probe has been
    ...    refused over IPv4 and is stuck on the black-holed IPv6 connect, lift
    ...    the IPv4 refusal, and require that same MicroShift process to reach
    ...    readiness without a restart.
    Stop MicroShift
    Probe Firewall Tables Should Be Absent
    Log Probe Preconditions
    ${cursor}=    Get Journal Cursor
    Black Hole IPv6 Connects To Kubelet Healthz
    Refuse IPv4 Connects To Kubelet Healthz
    Start MicroShift Without Waiting For Systemd Readiness
    ${pid}=    Wait Until Probe Is Stuck
    Allow IPv4 Connects To Kubelet Healthz
    Wait For MicroShift Service
    Same MicroShift Process Should Be Ready    ${pid}    ${cursor}
    Wait Until Keyword Succeeds    10x    3s
    ...    No Probe Connect Should Be Stuck In SYN-SENT    ${pid}
    [Teardown]    Run Keywords    Collect Probe Diagnostics On Failure
    ...    AND    Remove Probe Firewall Rules


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig

Teardown
    [Documentation]    Test suite teardown: remove any leftover rules and make
    ...    sure MicroShift is healthy again.
    Remove Probe Firewall Rules
    Restart MicroShift
    Logout MicroShift Host

Log Probe Preconditions
    [Documentation]    Record how localhost resolves and the kubelet healthz
    ...    listener, for diagnosing a failure.
    ${resolution}=    Command Should Work    getent ahosts localhost
    Log    ${resolution}
    ${listeners}    ${stderr}    ${rc}=    Command Execution
    ...    ss -Hltnp '( sport = :${KUBELET_HEALTHZ_PORT} )'
    Log    ${listeners}

Black Hole IPv6 Connects To Kubelet Healthz
    [Documentation]    Silently drop IPv6 packets to the kubelet healthz port on
    ...    ::1 so a connect hangs instead of being refused. The chain runs before
    ...    firewalld's output chain so the drop wins.
    Command Should Work    nft add table inet ${NFT_TABLE_V6}
    Command Should Work
    ...    nft 'add chain inet ${NFT_TABLE_V6} output { type filter hook output priority filter - 10; policy accept; }'
    Command Should Work
    ...    nft add rule inet ${NFT_TABLE_V6} output ip6 daddr ::1 tcp dport ${KUBELET_HEALTHZ_PORT} counter drop

Refuse IPv4 Connects To Kubelet Healthz
    [Documentation]    Reset IPv4 packets to the kubelet healthz port on
    ...    127.0.0.1 so the probe's connect is refused even if the kubelet is
    ...    already listening, which forces the fallback to ::1. The rule counts
    ...    the packets it rejects.
    Command Should Work    nft add table inet ${NFT_TABLE_V4}
    Command Should Work
    ...    nft 'add chain inet ${NFT_TABLE_V4} output { type filter hook output priority filter - 10; policy accept; }'
    Command Should Work
    ...    nft add rule inet ${NFT_TABLE_V4} output ip daddr 127.0.0.1 tcp dport ${KUBELET_HEALTHZ_PORT} counter reject with tcp reset

Allow IPv4 Connects To Kubelet Healthz
    [Documentation]    Lift the IPv4 refusal; the IPv6 black hole stays.
    Command Should Work    nft delete table inet ${NFT_TABLE_V4}

Probe Firewall Table Should Be Absent
    [Documentation]    Fail if the named test table exists.
    [Arguments]    ${table}
    ${tables}=    Command Should Work    nft list tables
    Should Not Contain    ${tables}    ${table}

Probe Firewall Tables Should Be Absent
    [Documentation]    Fail if either test table exists, so a leftover from an
    ...    earlier run is not mistaken for this test's fixture. The suite
    ...    assumes exclusive use of the host: teardown deletes these tables
    ...    whoever created them.
    Probe Firewall Table Should Be Absent    ${NFT_TABLE_V4}
    Probe Firewall Table Should Be Absent    ${NFT_TABLE_V6}

Remove Probe Firewall Table
    [Documentation]    Delete the named test table if it exists.
    [Arguments]    ${table}
    ${stdout}    ${stderr}    ${rc}=    Command Execution    nft list table inet ${table}
    IF    ${rc} == 0    Command Should Work    nft delete table inet ${table}

Remove Probe Firewall Rules
    [Documentation]    Delete both test tables and verify they are gone.
    Remove Probe Firewall Table    ${NFT_TABLE_V4}
    Remove Probe Firewall Table    ${NFT_TABLE_V6}
    Probe Firewall Tables Should Be Absent

IPv4 Refusal Should Have Been Hit
    [Documentation]    The IPv4 reject rule has counted at least one packet, so
    ...    the probe's IPv4 connect was refused.
    ${table}=    Command Should Work    nft list table inet ${NFT_TABLE_V4}
    Should Match Regexp    ${table}    counter packets [1-9]

Probe Should Be Stuck
    [Documentation]    The MicroShift main process holds a SYN-SENT socket to
    ...    the kubelet healthz port on ::1 and its IPv4 connect was refused.
    ...    Returns the main PID.
    ${pid}=    Get Systemd Setting    microshift.service    MainPID
    Should Not Be Equal As Integers    ${pid}    0
    ${sockets}=    Command Should Work
    ...    ss -Htnp state syn-sent '( dport = :${KUBELET_HEALTHZ_PORT} )'
    Should Match Regexp    ${sockets}
    ...    \\[::1\\]:${KUBELET_HEALTHZ_PORT}.*pid=${pid},
    IPv4 Refusal Should Have Been Hit
    RETURN    ${pid}

Wait Until Probe Is Stuck
    [Documentation]    Wait for the probe to be stuck as described in
    ...    "Probe Should Be Stuck" and return the main PID.
    ${pid}=    Wait Until Keyword Succeeds    30x    2s    Probe Should Be Stuck
    RETURN    ${pid}

Same MicroShift Process Should Be Ready
    [Documentation]    The main PID is unchanged, systemd counted no restart,
    ...    the kubelet probe succeeded and no service failed.
    [Arguments]    ${pid}    ${cursor}
    ${pid_now}=    Get Systemd Setting    microshift.service    MainPID
    Should Be Equal As Integers    ${pid_now}    ${pid}
    ...    msg=MicroShift restarted instead of recovering
    ${restarts}=    Get Systemd Setting    microshift.service    NRestarts
    Should Be Equal As Integers    ${restarts}    0
    Pattern Should Appear In Log Output    ${cursor}    kubelet is ready
    Pattern Should Not Appear In Log Output    ${cursor}    SERVICE FAILED    retries=0

No Probe Connect Should Be Stuck In SYN-SENT
    [Documentation]    The MicroShift main process holds no SYN-SENT socket to
    ...    the kubelet healthz port any more.
    [Arguments]    ${pid}
    ${sockets}=    Command Should Work
    ...    ss -Htnp state syn-sent '( dport = :${KUBELET_HEALTHZ_PORT} )'
    Should Not Match Regexp    ${sockets}
    ...    :${KUBELET_HEALTHZ_PORT}.*pid=${pid},

Collect Probe Diagnostics On Failure
    [Documentation]    On failure, record sockets, test tables and the service
    ...    state before the rules are removed.
    IF    "${TEST STATUS}" != "FAIL"    RETURN
    ${sockets}    ${stderr}    ${rc}=    Command Execution
    ...    ss -Htnpo '( sport = :${KUBELET_HEALTHZ_PORT} or dport = :${KUBELET_HEALTHZ_PORT} )'
    ${tables}    ${stderr}    ${rc}=    Command Execution    nft list ruleset
    Systemctl Print Service Status And Logs    microshift.service
