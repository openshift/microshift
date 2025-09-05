*** Settings ***
Documentation       Tests for various log messages we do or do not want.

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-process.resource
Resource            ../../resources/ostree-health.resource
Library             ../../resources/journalctl.py

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           restart    slow


*** Variables ***
${CURSOR}       ${EMPTY}    # The journal cursor before restarting MicroShift


*** Test Cases ***
Check Logs After Clean Start
    [Documentation]    Start from scratch, wait until initialized,
    ...    stop and check for errors.

    Cleanup MicroShift    --all    --keep-images
    ${cursor}=    Get Journal Cursor
    VAR    ${CURSOR}=    ${cursor}    scope=SUITE

    Enable MicroShift
    Start MicroShift
    Setup Kubeconfig

    Restart Greenboot And Wait For Success
    Stop MicroShift

    # Note: The 'forbidden' messages appear on clean startup.
    # Should Not Find Forbidden
    Should Not Find Cannot Patch Resource
    Services Should Not Timeout When Stopping
    Should Find Etcd Is Ready
    Should Find MicroShift Is Ready

Check Logs After Restart
    [Documentation]    Start again, wait until initialized,
    ...    stop and check for errors.

    ${cursor}=    Get Journal Cursor
    VAR    ${CURSOR}=    ${cursor}    scope=SUITE

    Start MicroShift
    Restart Greenboot And Wait For Success
    Stop MicroShift

    Should Not Find Forbidden
    Should Not Find Cannot Patch Resource
    Services Should Not Timeout When Stopping
    Should Find Etcd Is Ready
    Should Find MicroShift Is Ready


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host

Teardown
    [Documentation]    Test suite teardown
    Restart MicroShift
    Restart Greenboot And Wait For Success

    Logout MicroShift Host
    Remove Kubeconfig

Should Not Find Forbidden
    [Documentation]    Logs should not say "forbidden"
    Pattern Should Not Appear In Log Output    ${CURSOR}    forbidden

Should Not Find Cannot Patch Resource
    [Documentation]    Logs should not say "cannot patch resource"
    Pattern Should Not Appear In Log Output    ${CURSOR}    cannot patch resource

Services Should Not Timeout When Stopping
    [Documentation]    Logs should not say "Timed out waiting for services to stop"
    Pattern Should Not Appear In Log Output    ${CURSOR}    MICROSHIFT STOP TIMED OUT

Should Find Etcd Is Ready
    [Documentation]    Logs should say "etcd is ready"
    Pattern Should Appear In Log Output    ${CURSOR}    etcd is ready

Should Find MicroShift Is Ready
    [Documentation]    Logs should say "MICROSHIFT READY"
    Pattern Should Appear In Log Output    ${CURSOR}    MICROSHIFT READY
