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

${TARGET_REF}       ${EMPTY}


*** Test Cases ***
Upgrade
    [Documentation]    Performs an upgrade to given reference
    ...    and verifies if it was successful

    MicroShift 413 Should Not Have Upgrade Artifacts
    Wait Until Greenboot Health Check Exited
    System Should Be Healthy

    ${future_backup}=    Get Future Backup Name For Current Boot
    ${old_deploy_id}=    Get Booted Deployment Id

    Rebase System And Verify    ${TARGET_REF}
    ${new_deploy_id}=    Get Booted Deployment ID
    Should Not Be Equal As Strings    ${old_deploy_id}    ${new_deploy_id}

    Backup Should Exist    ${future_backup}


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    IF    "${TARGET_REF}"=="${EMPTY}"
        Fatal Error    TARGET_REF variable is required
    END
    Login MicroShift Host

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host
