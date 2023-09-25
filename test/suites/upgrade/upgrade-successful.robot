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

${TARGET_REF}       ${EMPTY}


*** Test Cases ***
Upgrade
    [Documentation]    Performs an upgrade to given reference
    ...    and verifies if it was successful, with SELinux validation

    MicroShift 413 Should Not Have Upgrade Artifacts
    Wait Until Greenboot Health Check Exited

    ${future_backup}=    Get Future Backup Name For Current Boot
    Deploy Commit Not Expecting A Rollback    ${TARGET_REF}
    Backup Should Exist    ${future_backup}

    # Upgrade from 4.13 is not officially supported, skipping due to failure in relabeling
    IF    "${future_backup}" != "4.13"
        Validate SELinux With Backup    ${future_backup}
    END


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Should Not Be Empty    ${TARGET_REF}    TARGET_REF variable is required
    Login MicroShift Host

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host
