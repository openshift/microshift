*** Settings ***
Documentation       Tests related to MicroShift upgradeability on OSTree-based systems

Resource            ../resources/common.resource
Resource            ../resources/ostree.resource
Library             Collections

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           ostree


*** Variables ***
${USHIFT_HOST}      ${EMPTY}
${USHIFT_USER}      ${EMPTY}


*** Test Cases ***
MicroShift Backs Up Data On Boot
    [Documentation]    Check if rebooting system will result in backing up
    ...    MicroShift data when system.json has 'backup' action

    Wait Until Greenboot Health Check Exited
    System Should Be Healthy
    Remove Existing Backup For Current Deployment
    ${future_backup}=    Get Future Backup Name For Current Boot

    Reboot MicroShift Host
    Wait For MicroShift Service

    Backup Should Exist    ${future_backup}

MicroShift Restores Data On Boot
    [Documentation]    Check if rebooting system will result in restoring
    ...    MicroShift data when system.json has 'restore' action

    Wait Until Greenboot Health Check Exited    # we don't want greenboot to overwrite system.json
    Remove Existing Backup For Current Deployment

    ${backup_name}=    Make Masquerading Backup
    Create Marker In Backup Dir    ${backup_name}

    Write Restore Action
    Reboot MicroShift Host
    Wait For MicroShift Service

    Marker Should Exist In Data Dir
    Remove Marker From Data Dir


*** Keywords ***
Setup
    [Documentation]    Test suite setup
    Check Required Env Variables
    Login MicroShift Host

Teardown
    [Documentation]    Test suite teardown
    Logout MicroShift Host

Make Masquerading Backup
    [Documentation]    Stops MicroShift and creates manual backup that
    ...    masquerades as automated one (by prefixing the backup name with deployment ID)
    Systemctl    stop    microshift.service

    ${deploy_id}=    Get Booted Deployment ID
    ${backup_name}=    Set Variable    ${deploy_id}_manual

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    microshift admin data backup --name "${backup_name}"
    ...    sudo=True    return_stderr=True    return_rc=True
    Should Be Equal As Integers    0    ${rc}

    RETURN    ${backup_name}

Write Restore Action
    [Documentation]    Writes a restore action to perform on next boot by running red script.
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    /etc/greenboot/red.d/40_microshift_set_restore.sh
    ...    sudo=True    return_stderr=True    return_rc=True
    Should Be Equal As Integers    0    ${rc}

Create Marker In Backup Dir
    [Documentation]    Creates a marker file in backup directory
    [Arguments]    ${backup_name}

    # create a marker that we expect to show up in data directory after restore
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    touch ${BACKUP_STORAGE}/${backup_name}/marker
    ...    sudo=True    return_stderr=True    return_rc=True
    Should Be Equal As Integers    0    ${rc}

Marker Should Exist In Data Dir
    [Documentation]    Checks if marker file exists in MicroShift data directory
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    stat ${DATA_DIR}/marker
    ...    sudo=True    return_stderr=True    return_rc=True
    Should Be Equal As Integers    0    ${rc}

Remove Marker From Data Dir
    [Documentation]    Removes marker file from MicroShift data directory
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    rm -fv ${DATA_DIR}/marker
    ...    sudo=True    return_stderr=True    return_rc=True
    Should Be Equal As Integers    0    ${rc}
