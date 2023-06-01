*** Settings ***
Documentation       Tests related to MicroShift upgradeability on OSTree-based systems

Resource            ../resources/common.resource
Resource            ../resources/ostree.resource
Resource            ../resources/systemd.resource
Resource            ../resources/microshift-process.resource
Library             Collections

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           ostree


*** Variables ***
${USHIFT_HOST}      ${EMPTY}
${USHIFT_USER}      ${EMPTY}


*** Test Cases ***
Rebooting Healthy System Should Result In Data Backup
    [Documentation]    Check if rebooting healthy system will result in backing up of MicroShift data

    Wait Until Greenboot Health Check Exited
    System Should Be Healthy
    Remove Existing Backup For Current Deployment

    Reboot MicroShift Host
    Wait For MicroShift Service

    Backup For Booted Deployment Should Exist


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host

Wait Until Greenboot Health Check Exited
    [Documentation]    Wait until greenboot healthchecks are done

    Wait Until Keyword Succeeds    5m    15s
    ...    Greenboot Health Check Exited

Greenboot Health Check Exited
    [Documentation]    Checks if greenboot-healthcheck finished running

    ${value}=    Get Systemd Setting    greenboot-healthcheck.service    SubState
    Should Be Equal As Strings    ${value}    exited

Remove Existing Backup For Current Deployment
    [Documentation]    Remove backups for current deployment

    ${path}=    Get Booted Deployment Backup Prefix Path
    Should Not Be Empty    ${path}

    ${rm_output}    ${rm_stderr}    ${rc}=    Execute Command
    ...    rm -rf ${path}*
    ...    sudo=True    return_stderr=True    return_rc=True
    Log    ${rm_stderr}
    Log    ${rm_output}
    Should Be Equal As Integers    0    ${rc}

Backup For Booted Deployment Should Exist
    [Documentation]    Asserts that backup for currently booted deployment exists
    ${exists}=    Does Backup For Booted Deployment Exist
    Should Be True    ${exists}

System Should Be Healthy
    [Documentation]    Asserts that persisted health information is "healthy"
    ${health}=    Get System Health
    Should Be Equal As Strings    healthy    ${health}
