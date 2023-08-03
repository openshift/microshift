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
${REASON}           ${EMPTY}


*** Test Cases ***
Staged Deployment Is Consistently Unhealthy
    [Documentation]    Verifies that whether the new deployment
    ...    cannot create a backup or is simply unhealthy,
    ...    system will rollback and become healthy.

    Wait For Healthy System
    ${backup}=    Get Future Backup Name For Current Boot

    TestAgent.Add Action For Next Deployment    every    ${REASON}
    Deploy Commit Expecting A Rollback    ${FAILING_REF}
    Wait For MicroShift Service
    Backup Should Exist    ${backup}
    Wait For Healthy System


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Should Not Be Empty    ${FAILING_REF}    FAILING_REF variable is required
    Should Not Be Empty    ${REASON}    REASON variable is required
    Login MicroShift Host

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host
