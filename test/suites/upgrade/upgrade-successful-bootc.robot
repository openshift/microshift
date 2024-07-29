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

    # The system before the upgrade should either by bootc or ostree
    ${is_bootc}=    Is System Bootc
    ${is_ostree}=    Is System OSTree
    ${cond}=    Evaluate    ${is_bootc} == ${FALSE} and ${is_ostree} == ${FALSE}
    Should Not Be True    ${cond}    The system should either be bootc or ostree

    Wait Until Greenboot Health Check Exited

    ${future_backup}=    Get Future Backup Name For Current Boot
    Deploy Bootc Commit Not Expecting A Rollback    ${BOOTC_REGISTRY}    ${TARGET_REF}
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
