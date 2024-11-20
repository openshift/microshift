*** Settings ***
Documentation       Tests related to how etcd is managed

Resource            ../../resources/common.resource
Resource            ../../resources/systemd.resource
Resource            ../../resources/microshift-config.resource
Resource            ../../resources/microshift-process.resource
Library             Collections

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           configuration    etcd    restart    slow


*** Variables ***
${ETCD_SYSTEMD_UNIT}    microshift-etcd.scope
${MEMLIMIT256}          SEPARATOR=\n
...                     ---
...                     etcd:
...                     \ \ memoryLimitMB: 256
${MEMLIMIT0}            SEPARATOR=\n
...                     ---
...                     etcd:
...                     \ \ memoryLimitMB: 0


*** Test Cases ***
Set MemoryHigh Limit Unlimited
    [Documentation]    The default configuration should not limit RAM
    ...
    ...    Since we cannot assume that the default configuration file is
    ...    being used, the test explicitly configures a '0' limit, which
    ...    is equivalent to not having any configuration at all.
    [Setup]    Setup With Custom Config    ${MEMLIMIT0}
    Expect MemoryHigh    infinity

Set MemoryHigh Limit 256MB
    [Documentation]    Set the memory limit for etcd to 128MB and ensure it takes effect
    [Setup]    Setup With Custom Config    ${MEMLIMIT256}
    # Expecting the setting to be 128 * 1024 * 1024
    Expect MemoryHigh    268435456
    [Teardown]    Restore Default Config


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig    # for readiness checks

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
    ${actual}=    Get Systemd Setting    microshift-etcd.scope    MemoryHigh
    # Using integer comparison is complicated here because sometimes
    # the returned or expected value is 'infinity'.
    Should Be Equal    ${expected}    ${actual}
