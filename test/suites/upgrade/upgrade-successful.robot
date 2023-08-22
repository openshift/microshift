*** Settings ***
Documentation       Tests related to upgrading MicroShift

Resource            ../../resources/common.resource
Resource            ../../resources/ostree.resource
Library             Collections

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           ostree


*** Variables ***
${USHIFT_HOST}      ${EMPTY}
${USHIFT_USER}      ${EMPTY}

${TARGET_REF}       ${EMPTY}


*** Test Cases ***
Upgrade
    [Documentation]    Performs an upgrade to given reference
    ...    and verifies if it was successful

    MicroShift 413 Should Not Have Upgrade Artifacts
    Wait For Healthy System

    ${future_backup}=    Get Future Backup Name For Current Boot
    Deploy Commit Not Expecting A Rollback    ${TARGET_REF}
    Backup Should Exist    ${future_backup}


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Should Not Be Empty    ${TARGET_REF}    TARGET_REF variable is required
    Login MicroShift Host

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host
