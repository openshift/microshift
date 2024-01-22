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
Should Not Find Forbidden
    [Documentation]    Logs should not say "forbidden"
    Pattern Should Not Appear In Log Output    ${CURSOR}    forbidden

Should Not Find Cannot Patch Resource
    [Documentation]    Logs should not say "cannot patch resource"
    Pattern Should Not Appear In Log Output    ${CURSOR}    cannot patch resource

Services Should Not Timeout When Stopping
    [Documentation]    Logs should not say "Timed out waiting for services to stop"
    [Tags]    robot:exclude
    Pattern Should Not Appear In Log Output    ${CURSOR}    Timed out waiting for services to stop

Should Find Etcd Is Ready
    [Documentation]    Logs should say "etcd is ready"
    [Tags]    robot:exclude
    Pattern Should Appear In Log Output    ${CURSOR}    etcd is ready

Should Find MicroShift Is Ready
    [Documentation]    Logs should say "MICROSHIFT READY"
    [Tags]    robot:exclude
    Pattern Should Appear In Log Output    ${CURSOR}    MICROSHIFT READY


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host
    Setup Kubeconfig    # for readiness checks
    # Save the journal cursor then restart MicroShift so we capture
    # the shutdown messages and startup messages.
    ${cursor}=    Get Journal Cursor
    Set Suite Variable    \${CURSOR}    ${cursor}
    Restart MicroShift
    Restart Greenboot And Wait For Success

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host
    Remove Kubeconfig
