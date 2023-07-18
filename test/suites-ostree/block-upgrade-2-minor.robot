*** Settings ***
Documentation       Tests related to upgrading MicroShift

Resource            ../resources/common.resource
Resource            ../resources/ostree.resource
Library             Collections

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           ostree


*** Variables ***
${USHIFT_HOST}                  ${EMPTY}
${USHIFT_USER}                  ${EMPTY}

${TOO_NEW_MICROSHIFT_REF}       ${EMPTY}


*** Test Cases ***
Upgrading MicroShift By Two Minor Versions Is Blocked
    [Documentation]    Test verifies if attempt to upgrade MicroShift from
    ...    X.Y-1 to X.Y+1 is blocked (skipping current X.Y).

    Wait For Healthy System
    ${initial_deploy_backup}=    Get Future Backup Name For Current Boot

    Deploy Commit Expecting A Rollback    ${TOO_NEW_MICROSHIFT_REF}

    Wait For Healthy System
    Backup Should Exist    ${initial_deploy_backup}
    Journal Should Have Information About Failed Version Comparison


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Should Not Be Empty    ${TOO_NEW_MICROSHIFT_REF}    FAILING_REF variable is required
    Login MicroShift Host

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host
