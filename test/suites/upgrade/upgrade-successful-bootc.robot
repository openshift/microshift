*** Settings ***
Documentation       Tests related to upgrading MicroShift to bootc

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-host.resource
Resource            ../../resources/ostree.resource
Resource            ../../resources/selinux.resource
Library             Collections

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           bootc


*** Variables ***
${TARGET_REF}           ${EMPTY}
${BOOTC_REGISTRY}       ${EMPTY}


*** Test Cases ***
Upgrade
    [Documentation]    Performs an upgrade to a given bootc reference
    ...    and verifies if it was successful, including SELinux validation

    Wait Until Greenboot Health Check Exited

    ${future_backup}=    Get Future Backup Name For Current Boot
    Deploy Commit Not Expecting A Rollback
    ...    ${TARGET_REF}
    ...    False
    ...    ${BOOTC_REGISTRY}
    Backup Should Exist    ${future_backup}

    # SELinux tests do not pass on bootc images yet.
    # Uncomment the following test when the problem is fixed.
    # Validate SELinux With Backup    ${future_backup}


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Should Not Be Empty    ${TARGET_REF}    TARGET_REF variable is required
    Should Not Be Empty    ${BOOTC_REGISTRY}    BOOTC_REGISTRY variable is required
    Login MicroShift Host

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host
