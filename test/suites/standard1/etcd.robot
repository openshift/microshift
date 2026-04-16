*** Settings ***
Documentation       Tests related to how etcd is managed

Resource            ../../resources/common.resource
Resource            ../../resources/systemd.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-process.resource
Library             Collections

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           etcd


*** Variables ***
${ETCD_SYSTEMD_UNIT}        microshift-etcd.scope
${ETCD_CA_CERT}             /var/lib/microshift/certs/etcd-signer/ca.crt
${ETCD_CLIENT_CERT}         /var/lib/microshift/certs/etcd-signer/apiserver-etcd-client/client.crt
${ETCD_CLIENT_KEY}          /var/lib/microshift/certs/etcd-signer/apiserver-etcd-client/client.key
${ETCD_ENDPOINT}            https://localhost:2379
${ETCDCTL_LOCAL_PATH}       ${EXECDIR}/../_output/bin/etcdctl
${ETCDCTL_BIN}              /tmp/etcdctl
${ETCDCTL_CMD}              ${ETCDCTL_BIN} --cacert=${ETCD_CA_CERT} --cert=${ETCD_CLIENT_CERT} --key=${ETCD_CLIENT_KEY} --endpoints=${ETCD_ENDPOINT}
${MEMLIMIT256}              SEPARATOR=\n
...                         ---
...                         etcd:
...                         \ \ memoryLimitMB: 256
${MEMLIMIT0}                SEPARATOR=\n
...                         ---
...                         etcd:
...                         \ \ memoryLimitMB: 0


*** Test Cases ***
Etcd Database Defragment Manually
    [Documentation]    Verify that etcd database can be manually defragmented
    ...    using etcdctl and the database size does not grow.
    ...    Creates artificial fragmentation first to ensure defrag has
    ...    meaningful work to do regardless of the initial DB state
    ...    (e.g. after etcd storage version migration).
    Create Etcd Fragmentation
    ${size_before}=    Get Etcd Database Size
    Command Should Work    ${ETCDCTL_CMD} defrag
    ${size_after}=    Get Etcd Database Size
    Should Be True    ${size_after} <= ${size_before}
    ...    msg=DB size after defrag (${size_after}) should not exceed size before (${size_before})
    [Teardown]    Command Should Work    ${ETCDCTL_CMD} alarm disarm

Etcd Runs As Transient Systemd Scope Unit
    [Documentation]    Verify that etcd runs as a transient systemd scope unit
    ...    managed by MicroShift with the expected systemd wiring.
    Systemctl Check Service SubState    ${ETCD_SYSTEMD_UNIT}    running
    ${transient}=    Get Systemd Setting    ${ETCD_SYSTEMD_UNIT}    Transient
    Should Be Equal As Strings    ${transient}    yes
    ${pid}=    MicroShift Etcd Process ID
    Should Not Be Empty    ${pid}
    ${binds_to}=    Get Systemd Setting    ${ETCD_SYSTEMD_UNIT}    BindsTo
    Should Contain    ${binds_to}    microshift.service
    ${before}=    Get Systemd Setting    ${ETCD_SYSTEMD_UNIT}    Before
    Should Contain    ${before}    microshift.service

Etcd Scope Follows MicroShift Lifecycle
    [Documentation]    Verify that etcd scope stops with MicroShift and restarts with it.
    [Tags]    restart    slow
    Stop MicroShift
    Wait Until Etcd Scope Is Inactive
    Start MicroShift
    Wait For MicroShift
    Systemctl Check Service SubState    ${ETCD_SYSTEMD_UNIT}    running
    [Teardown]    Run Keywords    Start MicroShift    AND    Wait For MicroShift

Set MemoryHigh Limit Unlimited
    [Documentation]    The default configuration should not limit RAM
    ...
    ...    Since we cannot assume that the default configuration file is
    ...    being used, the test explicitly configures a '0' limit, which
    ...    is equivalent to not having any configuration at all.
    [Tags]    configuration    restart    slow
    [Setup]    Setup With Custom Config    ${MEMLIMIT0}
    Expect MemoryHigh    infinity
    [Teardown]    Restore Default Config

Set MemoryHigh Limit 256MB
    [Documentation]    Set the memory limit for etcd to 256MB and ensure it takes effect
    [Tags]    configuration    restart    slow
    [Setup]    Setup With Custom Config    ${MEMLIMIT256}
    # Expecting the setting to be 256 * 1024 * 1024
    Expect MemoryHigh    268435456
    [Teardown]    Restore Default Config


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig    # for readiness checks
    Install Etcdctl

Teardown
    [Documentation]    Test suite teardown
    Restore Default Config
    Logout MicroShift Host
    Remove Kubeconfig

Restore Default Config
    [Documentation]    Remove any custom config and restart MicroShift
    Remove Drop In MicroShift Config    10-etcd
    Restart MicroShift

Setup With Custom Config
    [Documentation]    Install a custom config and restart MicroShift
    [Arguments]    ${config_content}
    Drop In MicroShift Config    ${config_content}    10-etcd
    Restart MicroShift

Expect MemoryHigh
    [Documentation]    Verify that the MemoryHigh setting for etcd matches the expected value
    [Arguments]    ${expected}
    ${actual}=    Get Systemd Setting    ${ETCD_SYSTEMD_UNIT}    MemoryHigh
    # Using integer comparison is complicated here because sometimes
    # the returned or expected value is 'infinity'.
    Should Be Equal    ${expected}    ${actual}

Etcd Scope Is Inactive
    [Documentation]    Check that the etcd scope unit is not active.
    ...    Transient scopes disappear when stopped, so is-active returns
    ...    "inactive" or an error.
    ${stdout}    ${rc}=    Execute Command
    ...    systemctl is-active ${ETCD_SYSTEMD_UNIT}
    ...    sudo=True    return_rc=True
    Should Match Regexp    ${stdout.strip()}    ^inactive$

Wait Until Etcd Scope Is Inactive
    [Documentation]    Wait for the etcd scope to become inactive
    Wait Until Keyword Succeeds    30x    5s
    ...    Etcd Scope Is Inactive

Install Etcdctl
    [Documentation]    Upload the pre-staged etcdctl binary to the remote host.
    ...    The binary is pre-downloaded with checksum verification by scripts/fetch_tools.sh.
    Put File    ${ETCDCTL_LOCAL_PATH}    ${ETCDCTL_BIN}

Create Etcd Fragmentation
    [Documentation]    Create artificial fragmentation in the etcd database by
    ...    writing and then deleting a set of keys. The deleted keys leave
    ...    dead space that defrag can reclaim, ensuring the defrag test
    ...    works even when the DB starts fully compacted.
    FOR    ${i}    IN RANGE    100
        Command Should Work    ${ETCDCTL_CMD} put /defrag-test/key-${i} "$(head -c 1024 /dev/urandom | base64 -w0)"
    END
    FOR    ${i}    IN RANGE    100
        Command Should Work    ${ETCDCTL_CMD} del /defrag-test/key-${i}
    END
    ${output}=    Command Should Work    ${ETCDCTL_CMD} endpoint status --write-out\=json
    ${revision}=    Command Should Work
    ...    printf '%s' '${output}' | python3 -c "import sys,json; print(json.load(sys.stdin)[0]['Status']['header']['revision'])"
    Command Should Work    ${ETCDCTL_CMD} compact ${revision}

Get Etcd Database Size
    [Documentation]    Return the current etcd database size in bytes
    ${output}=    Command Should Work    ${ETCDCTL_CMD} endpoint status --write-out\=json
    ${size}=    Command Should Work
    ...    printf '%s' '${output}' | python3 -c "import sys,json; print(json.load(sys.stdin)[0]['Status']['dbSize'])"
    RETURN    ${size}
