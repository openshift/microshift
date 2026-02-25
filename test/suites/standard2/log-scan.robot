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
Log Scan
    [Documentation]    Run log scan tests in a specific order.
    # Clean up and enable MicroShift to start from scratch
    Cleanup MicroShift    --all    --keep-images
    Enable MicroShift

    ${cursor}=    Get Journal Cursor
    VAR    ${CURSOR}=    ${cursor}    scope=SUITE
    # Start, stop and check logs after clean startup
    Start Stop And Check Logs    check_forbidden=False

    ${cursor}=    Get Journal Cursor
    VAR    ${CURSOR}=    ${cursor}    scope=SUITE
    # Restart, stop and check logs
    Start Stop And Check Logs


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host

Teardown
    [Documentation]    Test suite teardown
    Restart MicroShift
    Wait For MicroShift Healthcheck Success

    Logout MicroShift Host
    Remove Kubeconfig

Start Stop And Check Logs
    [Documentation]    Start, wait until initialized, stop and check for errors.
    [Arguments]    ${check_forbidden}=True
    Start MicroShift
    Setup Kubeconfig

    Wait For MicroShift Healthcheck Success
    Stop MicroShift

    # Note: The 'forbidden' messages appear on clean startup
    IF    ${check_forbidden}    Should Not Find Forbidden
    Should Not Find Cannot Patch Resource
    Services Should Not Timeout When Stopping
    Should Find Etcd Is Ready
    Should Find MicroShift Is Ready

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
