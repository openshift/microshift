*** Settings ***
Documentation       Tests related to upgrading MicroShift

Resource            ../../resources/common.resource
Resource            ../../resources/ostree.resource
Resource            ../../resources/selinux.resource
Library             Collections

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           ostree


*** Variables ***
${USHIFT_HOST}      ${EMPTY}
${USHIFT_USER}      ${EMPTY}
${BACKUP_NAME}      backup-name


*** Test Cases ***
Containers Should Not Have Access To Container Var Lib Labels
    [Documentation]    Performs a check to make sure containers can not access
    ...    files or folders that are labeled with container var lib as well as the
    ...    generated backup file

    ${result}=    Run Container File Access    "${BACKUP_STORAGE}/${BACKUP_NAME}/version"
    Should Be Empty    ${result}

Folders Should Have Expected Fcontext Types
    [Documentation]    Performs a check to make sure the folders created during rpm install
    ...    have the expected fcontext values

    ${result}=    Run Fcontext Types Check
    Should Be Empty    ${result}

Semanage Fcontext Should Have Combined List of OCP and MicroShift Rules
    [Documentation]    Validates that the fcontext data is the combined set for
    ...    OCP and MicroShift

    ${result}=    Get Current Fcontext List
    ${expected}=    Get Expected Fcontext List
    Lists Should Be Equal    ${result}    ${expected}

Audit Log Should Be Empty For MicroShift
    [Documentation]    Checks that no permission denials have occured during running MicroShift

    ${result}=    Get Denial Audit Log List
    Should Be Empty    ${result}


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host
    Wait For Healthy System
    Create Backup    ${BACKUP_NAME}

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host
