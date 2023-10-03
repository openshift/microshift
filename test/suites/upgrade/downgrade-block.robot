*** Settings ***
Documentation       Tests related to upgrading MicroShift

Resource            ../../resources/common.resource
Resource            ../../resources/ostree.resource
Library             Collections

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           ostree


*** Variables ***
${USHIFT_HOST}              ${EMPTY}
${USHIFT_USER}              ${EMPTY}

${OLDER_MICROSHIFT_REF}     ${EMPTY}


*** Test Cases ***
Downgrade Is Blocked
    [Documentation]    Verifies that staging a new deployment featuring
    ...    MicroShift "older" than existing data is blocked
    ...    and results in system rolling back.

    ${initial_deploy_backup}=    Get Future Backup Name For Current Boot

    Deploy Commit Expecting A Rollback    ${OLDER_MICROSHIFT_REF}    write_agent_cfg=False
    Wait Until Greenboot Health Check Exited
    Backup Should Exist    ${initial_deploy_backup}
    Journal Should Have Information About Failed Version Comparison


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Should Not Be Empty    ${OLDER_MICROSHIFT_REF}    FAKE_NEXT_MINOR_REF variable is required
    Login MicroShift Host
    Wait Until Greenboot Health Check Exited

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host
