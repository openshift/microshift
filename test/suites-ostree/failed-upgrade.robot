*** Settings ***
Documentation       Tests related to upgrading MicroShift

Resource            ../resources/common.resource
Resource            ../resources/ostree.resource
Library             Collections

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           ostree


*** Variables ***
${USHIFT_HOST}      ${EMPTY}
${USHIFT_USER}      ${EMPTY}

${FAILING_REF}      ${EMPTY}


*** Test Cases ***
Staged Deployment Consistently Fails To Back Up The Data
    [Documentation]    Verifies that instructions to back up the data
    ...    ("healthy" system) is not lost if staged deployment fails
    ...    to back up and it is performed after rolling back.

    Wait For Healthy System
    ${backup}=    Get Future Backup Name For Current Boot

    TestAgent.Add Action For Next Deployment    every    prevent_backup
    Deploy Commit Expecting A Rollback    ${FAILING_REF}
    Wait For MicroShift Service
    Backup Should Exist    ${backup}


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Should Not Be Empty    ${FAILING_REF}    FAILING_REF variable is required
    Login MicroShift Host

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host
