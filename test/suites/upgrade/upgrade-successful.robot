*** Settings ***
Documentation       Tests related to upgrading MicroShift

Resource            ../../resources/common.resource
Resource            ../../resources/microshift-host.resource
Resource            ../../resources/ostree.resource
Resource            ../../resources/oc.resource
Resource            ../../resources/selinux.resource
Library             Collections

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           ostree


*** Variables ***
${TARGET_REF}           ${EMPTY}
${BOOTC_REGISTRY}       ${EMPTY}


*** Test Cases ***
Upgrade
    [Documentation]    Performs an upgrade to given reference
    ...    and verifies if it was successful, with SELinux validation

    Wait Until Greenboot Health Check Exited

    ${future_backup}=    Get Future Backup Name For Current Boot
    Deploy Commit Not Expecting A Rollback
    ...    ${TARGET_REF}
    ...    False
    ...    ${BOOTC_REGISTRY}
    Backup Should Exist    ${future_backup}

    # SELinux tests do not pass on bootc images yet.
    # Run the following test on bootc systems when the problem is fixed.
    ${is_bootc}=    Is System Bootc
    IF    ${is_bootc} == ${FALSE}
        Validate SELinux With Backup    ${future_backup}
    END


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Should Not Be Empty    ${TARGET_REF}    TARGET_REF variable is required
    Login MicroShift Host
    Setup Kubeconfig

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host
