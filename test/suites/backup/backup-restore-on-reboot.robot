*** Settings ***
Documentation       Tests related to MicroShift upgradeability on OSTree-based systems

Resource            ../../resources/common.resource
Resource            ../../resources/ostree.resource
Library             Collections

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           ostree


*** Variables ***
${USHIFT_HOST}      ${EMPTY}
${USHIFT_USER}      ${EMPTY}


*** Test Cases ***
Rebooting System Should Result In Data Backup
    [Documentation]    Check if rebooting system will result in MicroShift creating backup

    Wait For MicroShift Service
    ${future_backup}=    Get Future Backup Name For Current Boot

    Reboot MicroShift Host
    Wait For MicroShift Service

    Backup Should Exist    ${future_backup}

Existence Of Restore File Should Result In Restoring Data
    [Documentation]    Check if rebooting unhealthy system will result
    ...    restoring MicroShift data from a backup

    Wait For MicroShift Service
    Remove Existing Backup For Current Deployment

    ${backup_name}=    Make Masquerading Backup

    Override Deployment ID In Version File
    Create Restore Marker File    # to trigger restore after reboot
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
    VAR    ${backup_name}=    ${deploy_id}_manual

    Create Backup    ${backup_name}    ${TRUE}

    RETURN    ${backup_name}

Create Restore Marker File
    [Documentation]    Creates a special file which existence of MicroShift will
    ...    interpret as request to restore a backup.

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    touch /var/lib/microshift-backups/restore
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

Override Deployment ID In Version File
    [Documentation]    Changes deployment_id in the version file to bypass
    ...    check that blocks restore when current deployment matches
    ...    deployment saved in the file.

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    jq '.deployment_id="test"' ${DATA_DIR}/version > /tmp/microshift-version.new
    ...    sudo=True
    ...    return_stderr=True
    ...    return_rc=True
    Should Be Equal As Integers    0    ${rc}

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    sudo mv /tmp/microshift-version.new ${DATA_DIR}/version
    ...    sudo=True
    ...    return_stderr=True
    ...    return_rc=True
    Should Be Equal As Integers    0    ${rc}
