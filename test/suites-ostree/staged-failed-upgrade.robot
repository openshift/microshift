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
First Failed Staged Upgrade Should Restore Backup
    [Documentation]    Staged deployment is unhealthy on first boot. After auto-greenboot-reboot MicroShift
    ...    should restore rollback deployment's backup to reattempt the upgrade from healthy data.

    Wait For Healthy System

    ${initial_backup}=    Get Future Backup Name For Current Boot
    Create Backup With Marker    ${initial_backup}

    TestAgent.Add Action For Next Deployment    1    fail_greenboot

    # We don't expect a rollback here since after the first failure
    # things should continue as desired. This test validates the
    # restore function is working when the first boot is unhealthy.
    Deploy Commit Not Expecting A Rollback    ${TARGET_REF}    ${TRUE}

    Wait For Healthy System

    Validate Backup Is Restored    ${initial_backup}
    Expected Boot Count    3


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

Create Backup With Marker
    [Documentation]    Stops MicroShift and creates manual backup that
    ...    masquerades as automated one and creates marker file for validating restore.
    [Arguments]    ${backup_name}
    Systemctl    stop    microshift.service

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    microshift backup --name "${backup_name}"
    ...    sudo=True    return_stderr=True    return_rc=True

    # create a marker that we expect to show up in data directory after restore
    ${mark_stdout}    ${mark_stderr}    ${mark_rc}=    Execute Command
    ...    touch ${BACKUP_STORAGE}/${backup_name}/marker
    ...    sudo=True    return_stderr=True    return_rc=True

    Should Be Equal As Integers    0    ${rc}
    Should Be Equal As Integers    0    ${mark_rc}

Validate Backup Is Restored
    [Documentation]    Validate that the desired backup exists and is the only one that exists
    [Arguments]    ${backup_name}

    Backup Should Exist    ${backup_name}

    ${stdout}    ${rc}=    Execute Command
    ...    cd ${BACKUP_STORAGE} && ls -d1 rhel-*
    ...    sudo=False    return_rc=True

    ${mark_stdout}    ${mark_stderr}    ${mark_rc}=    Execute Command
    ...    stat ${DATA_DIR}/marker
    ...    sudo=True    return_stderr=True    return_rc=True

    Should Be Equal As Strings    ${stdout}    ${backup_name}
    Should Be Equal As Integers    0    ${mark_rc}

Expected Boot Count
    [Documentation]    Validate that the host rebooted only the specified number of times
    [Arguments]    ${reboot_count}

    ${stdout}    ${rc}=    Execute Command
    ...    journalctl --list-boots --quiet | wc -l
    ...    sudo=True    return_rc=True

    Should Be Equal As Integers    ${reboot_count}    ${stdout}
