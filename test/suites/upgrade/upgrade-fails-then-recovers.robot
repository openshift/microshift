*** Settings ***
Documentation       Tests related to upgrading MicroShift

Resource            ../../resources/common.resource
Resource            ../../resources/ostree.resource
Library             Collections

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           ostree


*** Variables ***
${TARGET_REF}           ${EMPTY}
${BOOTC_REGISTRY}       ${EMPTY}


*** Test Cases ***
First Failed Staged Upgrade Should Not Restore Backup
    [Documentation]    Staged deployment is unhealthy on first boot.
    ...    After auto-greenboot-reboot MicroShift should not restore
    ...    rollback deployment's backup, instead it should just carry on with
    ...    the existing data.

    Wait Until Greenboot Health Check Exited

    ${initial_backup}=    Get Future Backup Name For Current Boot
    Create Backup    ${initial_backup}    ${TRUE}

    TestAgent.Add Action For Next Deployment    1    fail_greenboot

    # We don't expect a rollback here since after the first failure
    # MicroShift should continue with business as usual
    Deploy Commit Not Expecting A Rollback
    ...    ${TARGET_REF}
    ...    ${TRUE}
    ...    ${BOOTC_REGISTRY}

    Wait Until Greenboot Health Check Exited

    Validate Backup Is Not Restored    ${initial_backup}
    Expected Boot Count    3


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Should Not Be Empty    ${TARGET_REF}    TARGET_REF variable is required
    Login MicroShift Host

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host

Validate Backup Is Not Restored
    [Documentation]    Validate that two backups exist - one for rollback
    ...    deployment (with marker), and second for current deployment
    ...    (despite failing greenboot on first boot).
    [Arguments]    ${backup_name}

    Backup Should Exist    ${backup_name}

    ${stdout}=    Execute Command
    ...    find "${BACKUP_STORAGE}" -maxdepth 1 -type d "(" -name "rhel-*" -o -name "default-*" ")" | wc -l
    ...    sudo=False    return_rc=False

    ${mark_stdout}    ${mark_stderr}    ${mark_rc}=    Execute Command
    ...    stat ${DATA_DIR}/marker
    ...    sudo=True    return_stderr=True    return_rc=True

    Should Be Equal As Strings    ${stdout}    2
    Should Be Equal As Integers    1    ${mark_rc}
