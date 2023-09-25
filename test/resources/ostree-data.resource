*** Settings ***
Documentation       Keywords for OSTree-based systems

Resource            systemd.resource
Resource            microshift-process.resource
Library             libostree.py


*** Variables ***
${DATA_DIR}             /var/lib/microshift
${BACKUP_STORAGE}       /var/lib/microshift-backups


*** Keywords ***
Get Future Backup Name For Current Boot
    [Documentation]    Provides a name for future backup:
    ...    "4.13" for MicroShift 4.13, otherwise deployment ID + boot ID

    # When system is upgrading from 4.13, there is no metadata regarding previous boot,
    # therefore, when naming backup for existing data, "4.13" will be used as deployment ID,
    # i.e. it should reside in `/var/lib/microshift-backups/4.13`
    ${version}=    MicroShift Version
    IF    ${version.minor} == 13    RETURN    4.13

    ${deploy_id}=    Get Booted Deployment ID
    ${boot_id}=    Get Current Boot Id
    ${backup_name}=    Catenate    SEPARATOR=_    ${deploy_id}    ${boot_id}
    RETURN    ${backup_name}

Create Backup
    [Documentation]    Stops MicroShift and creates manual backup
    ...    and optional marker file for validating.
    [Arguments]    ${backup_name}    ${create_marker_file}=${FALSE}
    Systemctl    stop    microshift.service

    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    microshift backup "${BACKUP_STORAGE}/${backup_name}"
    ...    sudo=True    return_stderr=True    return_rc=True
    Should Be Equal As Integers    0    ${rc}

    IF    ${create_marker_file}
        # create a marker that we expect to show up in data directory after restore
        ${mark_stdout}    ${mark_stderr}    ${mark_rc}=    Execute Command
        ...    touch ${BACKUP_STORAGE}/${backup_name}/marker
        ...    sudo=True    return_stderr=True    return_rc=True
        Should Be Equal As Integers    0    ${mark_rc}
    END

Backup Should Exist
    [Documentation]    Checks if backup identified by deployment ID
    ...    and (optionally) boot ID exists
    [Arguments]    ${deployment_id}    ${boot_id}=${EMPTY}
    ${exists}=    Does Backup Exist    ${deployment_id}    ${boot_id}
    Should Be True    ${exists}

Remove Existing Backup For Current Deployment
    [Documentation]    Remove any existing backup for currently running deployment
    ${deploy_id}=    Get Booted Deployment Id
    Remove Backups For Deployment    ${deploy_id}

MicroShift 413 Should Not Have Upgrade Artifacts
    [Documentation]    Verifies that host running MicroShift 4.13
    ...    does not have upgrade related artifacts.

    ${version}=    MicroShift Version
    IF    ${version.minor} == 13
        Path Should Not Exist    ${BACKUP_STORAGE}
        Path Should Not Exist    /var/lib/microshift/version
    END
