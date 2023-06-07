*** Settings ***
Documentation       Tests related to MicroShift upgradeability on OSTree-based systems

Resource            ../resources/common.resource
Resource            ../resources/ostree.resource
Resource            ../resources/systemd.resource
Resource            ../resources/microshift-process.resource
Library             Collections

Suite Setup         Setup
Suite Teardown      Teardown

Test Tags           ostree


*** Variables ***
${USHIFT_HOST}      ${EMPTY}
${USHIFT_USER}      ${EMPTY}


*** Test Cases ***
Rebooting Healthy System Should Result In Data Backup
    [Documentation]    Check if rebooting healthy system will result in backing up of MicroShift data

    Wait Until Greenboot Health Check Exited
    System Should Be Healthy
    Remove Existing Backup For Current Deployment

    Reboot MicroShift Host
    Wait For MicroShift Service

    Backup For Booted Deployment Should Exist

Rebooting Unhealthy System Should Result In Restoring Data From A Backup
    [Documentation]    Check if rebooting unhealthy system will result
    ...    restoring MicroShift data from a backup

    Wait Until Greenboot Health Check Exited    # we don't want greenboot to overwrite health.json
    Remove Existing Backup For Current Deployment

    ${backup_name}=    Make Masquerading Backup
    Create Marker In Backup Dir    ${backup_name}

    Mark System As Unhealthy    # to trigger restore after reboot
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

Wait Until Greenboot Health Check Exited
    [Documentation]    Wait until greenboot healthchecks are done

    Wait Until Keyword Succeeds    5m    15s
    ...    Greenboot Health Check Exited

Greenboot Health Check Exited
    [Documentation]    Checks if greenboot-healthcheck finished running

    ${value}=    Get Systemd Setting    greenboot-healthcheck.service    SubState
    Should Be Equal As Strings    ${value}    exited

Remove Existing Backup For Current Deployment
    [Documentation]    Remove backups for current deployment

    ${path}=    Get Booted Deployment Backup Prefix Path
    Should Not Be Empty    ${path}

    ${rm_output}    ${rm_stderr}    ${rc}=    Execute Command
    ...    rm -rf ${path}*
    ...    sudo=True    return_stderr=True    return_rc=True
    Log    ${rm_stderr}
    Log    ${rm_output}
    Should Be Equal As Integers    0    ${rc}

Backup For Booted Deployment Should Exist
    [Documentation]    Asserts that backup for currently booted deployment exists
    ${exists}=    Does Backup For Booted Deployment Exist
    Should Be True    ${exists}

System Should Be Healthy
    [Documentation]    Asserts that persisted health information is "healthy"
    ${health}=    Get System Health
    Should Be Equal As Strings    healthy    ${health}

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

Mark System As Unhealthy
    [Documentation]    Marks systems as unhealthy by executing microshift's red script
    ${stdout}    ${stderr}    ${rc}=    Execute Command
    ...    /etc/greenboot/red.d/40_microshift_set_unhealthy.sh
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
